import { env } from "cloudflare:workers";
import matter from "gray-matter";
import MarkdownIt from "markdown-it";
import texmath from "markdown-it-texmath";
import katex from "katex";

export interface Note {
	id: string;
	title: string;
	description: string;
	createdDate: Date;
	updatedDate?: Date;
	html: string;
	headings: Heading[];
}

export interface Heading {
	level: number;
	text: string;
	slug: string;
}

export interface SubjectNode {
	name: string;
	fullPath: string;
	noteCount: number;
	note?: Note;
	children: SubjectNode[];
}

const R2_PUBLIC_ASSETS_URL = "https://sqotion-storage.sqot0.my.id";

// Optional storage folder prefix (e.g. "notes") that wraps all keys in the
// bucket. Notes live at "{prefix}/..."..
const STORAGE_PREFIX = (
	import.meta.env.S3_PREFIX ?? ""
).replace(/^\/+|\/+$/g, "");
const STORAGE_PREFIX_DIR = STORAGE_PREFIX ? STORAGE_PREFIX + "/" : "";

const md = new MarkdownIt({
	html: true,
});

md.use(texmath, {
	engine: katex,
	delimiters: "dollars",
});

// Custom heading renderer to add IDs for TOC linking
md.renderer.rules.heading_open = (tokens, idx, options, _env, self) => {
	const token = tokens[idx];
	const nextToken = tokens[idx + 1];

	if (nextToken && nextToken.type === "inline") {
		const slug = slugify(getInlineText(nextToken));
		token.attrSet("id", slug);
	}

	return self.renderToken(tokens, idx, options);
};

// Custom image renderer to prepend R2 URL
const defaultImageRenderer =
	md.renderer.rules.image ?? md.renderer.renderToken.bind(md.renderer);
md.renderer.rules.image = (tokens, idx, options, _env, self) => {
	const token = tokens[idx];
	const srcIndex = token.attrIndex("src");
	if (srcIndex >= 0) {
		const [name, value] = token.attrs![srcIndex];
		token.attrs![srcIndex] = [name, `${R2_PUBLIC_ASSETS_URL}/${STORAGE_PREFIX_DIR}${value}`];
	}
	return defaultImageRenderer(tokens, idx, options, _env, self);
};

function slugify(text: string): string {
	return text
		.toLowerCase()
		.replace(/[^\w\s-]/g, "")
		.replace(/\s+/g, "-")
		.replace(/-+/g, "-")
		.trim();
}

function getInlineText(token: any): string {
	if (typeof token === "string") return token;
	if (token.content) return token.content;
	if (token.children) {
		return token.children.map(getInlineText).join("");
	}
	return "";
}

function extractHeadings(html: string): Heading[] {
	const headingRegex = /<h([2-4])\s+id="([^"]*)"[^>]*>(.*?)<\/h[2-4]>/gi;
	const headings: Heading[] = [];
	let match: RegExpExecArray | null;

	while ((match = headingRegex.exec(html)) !== null) {
		headings.push({
			level: parseInt(match[1]),
			text: match[3].replace(/<[^>]*>/g, ""),
			slug: match[2],
		});
	}

	return headings;
}

async function parseNote(key: string, object: R2ObjectBody): Promise<Note> {
	const textContent = await object.text();
	const { data, content } = matter(textContent);
	const html = md.render(content);

	return {
		id: key.slice(STORAGE_PREFIX_DIR.length).replace(/\.md$/, ""),
		title: typeof data.title === "string" ? data.title : "",
		description: typeof data.description === "string" ? data.description : "",
		createdDate: object.uploaded,
		updatedDate: data.updatedDate ? new Date(data.updatedDate) : undefined,
		html,
		headings: extractHeadings(html),
	};
}

function isMdKey(key: string): boolean {
	const rel = key.slice(STORAGE_PREFIX_DIR.length);
	const firstSegment = rel.split("/")[0] ?? "";
	// Ignore the Assets folder entirely.
	if (firstSegment.toLowerCase() === "assets") {
		return false;
	}
	// Only Markdown files are notes. Everything else (images, etc.) is ignored.
	return rel.endsWith(".md");
}

async function listObjects(folderPath?: string): Promise<R2ObjectBody[]> {
	const bodies: R2ObjectBody[] = [];
	let cursor: string | undefined;
	const r2Prefix = folderPath
		? `${STORAGE_PREFIX_DIR}${folderPath.replace(/\/+$/, "")}/`
		: STORAGE_PREFIX_DIR;

	do {
		const result = await env.STORAGE.list({ prefix: r2Prefix, cursor });
		bodies.push(
			...(await Promise.all(
				result.objects
					.filter((obj) => isMdKey(obj.key))
					.map(async (obj) => {
						const object = await env.STORAGE.get(obj.key);
						if (!object) {
							throw new Error(`Failed to retrieve note: ${obj.key}`);
						}
						return object;
					}),
			)),
		);
		// List() paginates at 1000 objects; follow the cursor when present.
		cursor = result.truncated ? result.cursor : undefined;
	} while (cursor);

	return bodies;
}

export async function getNotes(): Promise<Note[]> {
	const bodies = await listObjects();
	return Promise.all(bodies.map((body) => parseNote(body.key, body)));
}

/**
 * Returns all notes that live inside the given folder path (recursively),
 * e.g. getNotesByPath("math") returns notes with ids like "math/algebra".
 */
export async function getNotesByPath(path: string): Promise<Note[]> {
	const normalized = path.replace(/\/+$/, "");
	const bodies = await listObjects(normalized || undefined);
	return Promise.all(bodies.map((body) => parseNote(body.key, body)));
}

export async function getNote(id: string): Promise<Note | null> {
	const key = `${STORAGE_PREFIX_DIR}${id.replace(/\.md$/, "")}.md`;
	const object = await env.STORAGE.get(key);
	if (!object) {
		return null;
	}
	return parseNote(key, object);
}

/**
 * Deterministically classify a slug as either a note (leaf) or a folder
 * based on the note ids that exist.
 */
export function classifyPath(
	slug: string,
	noteIds: string[],
): "note" | "folder" | "none" {
	const normalized = slug.replace(/\/+$/, "").replace(/\.md$/, "");
	const idSet = new Set(noteIds);

	if (idSet.has(normalized)) {
		return "note";
	}
	if (noteIds.some((id) => id.startsWith(normalized + "/"))) {
		return "folder";
	}
	return "none";
}

/**
 * Build a nested subject/folder tree from a flat list of notes.
 * A leaf node represents a single note; internal nodes represent folders.
 */
export function buildSubjectTree(notes: Note[]): SubjectNode[] {
	const roots = new Map<string, SubjectNode>();
	const byPath = new Map<string, SubjectNode>();

	for (const note of notes) {
		const relId = note.id.startsWith(STORAGE_PREFIX_DIR)
			? note.id.slice(STORAGE_PREFIX_DIR.length)
			: note.id;
		const segments = relId.split("/").filter(Boolean);
		let current: SubjectNode | undefined;
		let prefix = "";

		for (let i = 0; i < segments.length; i++) {
			prefix = prefix ? `${prefix}/${segments[i]}` : segments[i];
			const isLeaf = i === segments.length - 1;

			let node = byPath.get(prefix);
			if (!node) {
				node = {
					name: segments[i],
					fullPath: prefix,
					noteCount: 0,
					children: [],
				};
				byPath.set(prefix, node);

				if (current) {
					current.children.push(node);
				} else {
					roots.set(prefix, node);
				}
			}

			if (isLeaf) {
				node.note = note;
			}

			current = node;
		}
	}

	// Accumulate note counts up the tree.
	const accumulate = (node: SubjectNode): number => {
		node.noteCount =
			(node.note ? 1 : 0) +
			node.children.reduce((sum, child) => sum + accumulate(child), 0);
		return node.noteCount;
	};

	const result = [...roots.values()];
	result.forEach(accumulate);
	return result;
}
