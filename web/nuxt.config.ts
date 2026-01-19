// https://nuxt.com/docs/api/configuration/nuxt-config
import tailwindcss from "@tailwindcss/vite";

export default defineNuxtConfig({
  compatibilityDate: "2025-07-15",
  devtools: { enabled: true },
  css: ["~/assets/css/tailwind.css"],
  ssr: true, // Enable SSR for prerendering SEO meta tags

  nitro: {
    output: {
      publicDir: "dist",
    },
    prerender: {
      crawlLinks: true,
      routes: ["/"],
    },
  },

  app: {
    head: {
      title: "hyperwhisper",
      meta: [
        { charset: "utf-8" },
        { name: "viewport", content: "width, initial-scale=1" },
        {
          name: "description",
          content: "Type 3x faster, without lifting a finger",
        },
        // Open Graph
        {
          property: "og:type",
          content: "website",
        },
        {
          property: "og:title",
          content: "hyperwhisper",
        },
        {
          property: "og:description",
          content: "Type 3x faster, without lifting a finger",
        },
        {
          property: "og:image",
          content: "https://hyperwhisper.dev/og-image.png",
        },
        {
          property: "og:url",
          content: "https://hyperwhisper.dev",
        },
        {
          property: "og:site_name",
          content: "hyperwhisper",
        },
        // Twitter Card
        {
          name: "twitter:card",
          content: "summary_large_image",
        },
        {
          name: "twitter:title",
          content: "hyperwhisper",
        },
        {
          name: "twitter:description",
          content: "Type 3x faster, without lifting a finger",
        },
        {
          name: "twitter:image",
          content: "https://hyperwhisper.dev/og-image.png",
        },
      ],
      link: [{ rel: "icon", type: "image/x-icon", href: "/favicon.ico" }],
    },
  },

  vite: {
    plugins: [tailwindcss()],
    server: {
      watch: {
        usePolling: true,
        interval: 1000,
      },
    },
  },

  modules: ["shadcn-nuxt", "@nuxtjs/color-mode"],

  colorMode: {
    classSuffix: "",
    preference: "dark",
    fallback: "dark",
  },

  shadcn: {
    /**
     * Prefix for all the imported component.
     * @default "Ui"
     */
    prefix: "",
    /**
     * Directory that the component lives in.
     * Will respect the Nuxt aliases.
     * @link https://nuxt.com/docs/api/nuxt-config#alias
     * @default "@/components/ui"
     */
    componentDir: "@/components/ui",
  },
});

