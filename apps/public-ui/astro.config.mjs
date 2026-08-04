// @ts-check
import { defineConfig, fontProviders } from "astro/config";

import cloudflare from "@astrojs/cloudflare";

import tailwindcss from "@tailwindcss/vite";
import pagefind from "astro-pagefind";

// https://astro.build/config
export default defineConfig({
	adapter: cloudflare(),
	integrations: [pagefind()],

	vite: {
		plugins: [tailwindcss()],
	},

	fonts: [
		{
			provider: fontProviders.google(),
			name: "Inter",
			cssVariable: "--font-inter",
		},
	],
});
