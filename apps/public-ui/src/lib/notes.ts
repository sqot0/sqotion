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
}

const R2_PUBLIC_ASSETS_URL =
  "https://sqotion-storage.sqot0.my.id";

const md = new MarkdownIt({
  html: true,
});

md.use(texmath, {
  engine: katex,
  delimiters: "dollars",
});

const defaultImageRenderer =
  md.renderer.rules.image ?? md.renderer.renderToken.bind(md.renderer);
md.renderer.rules.image = (tokens, idx, options, env, self) => {
  const token = tokens[idx];
  const srcIndex = token.attrIndex("src");
  if (srcIndex >= 0) {
    const [name, value] = token.attrs![srcIndex];
    token.attrs![srcIndex] = [name, `${R2_PUBLIC_ASSETS_URL}/${value}`];
  }
  return defaultImageRenderer(tokens, idx, options, env, self);
};

async function parseNote(key: string, object: R2ObjectBody): Promise<Note> {
  const textContent = await object.text();
  const { data, content } = matter(textContent);

  return {
    id: key.replace(/\.md$/, ""),
    title: typeof data.title === "string" ? data.title : "",
    description: typeof data.description === "string" ? data.description : "",
    createdDate: object.uploaded,
    updatedDate: data.updatedDate ? new Date(data.updatedDate) : undefined,
    html: md.render(content),
  };
}

export async function getNotes(): Promise<Note[]> {
  const { objects } = await env.STORAGE.list();

  return Promise.all(
    objects
      .filter((obj) => obj.key.endsWith(".md"))
      .map(async (obj) => {
        const object = await env.STORAGE.get(obj.key);
        if (!object) {
          throw new Error(`Failed to retrieve note: ${obj.key}`);
        }
        return parseNote(obj.key, object);
      }),
  );
}

export async function getNote(id: string): Promise<Note | null> {
  const key = `${id.replace(/\.md$/, "")}.md`;
  const object = await env.STORAGE.get(key);
  if (!object) {
    return null;
  }
  return parseNote(key, object);
}
