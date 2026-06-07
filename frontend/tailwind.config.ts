import type { Config } from "tailwindcss";

export default {
  content: [
    "./components/**/*.{vue,ts}",
    "./layouts/**/*.vue",
    "./pages/**/*.vue",
    "./plugins/**/*.ts",
    "./app.vue",
    "./error.vue",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['"Inter"', "ui-sans-serif", "system-ui", "sans-serif"],
      },
      colors: {
        // Donezo green palette (matches legacy design tokens)
        brand: {
          DEFAULT: "#1f7a4d",
          dark: "#0f5f3a",
          light: "#d4edda",
          bg: "#e8f5ee",
          soft: "#f0f9f4",
        },
        ink: {
          DEFAULT: "#1a1a2e",
          medium: "#4a5568",
          soft: "#8b9dc3",
        },
        page: "#f1f4f8",
        border: {
          soft: "#eef2f6",
        },
      },
      borderRadius: {
        xl: "16px",
        "2xl": "20px",
        "3xl": "24px",
      },
      boxShadow: {
        card: "0 1px 3px rgba(0,0,0,0.04)",
        hover: "0 8px 24px -8px rgba(0,0,0,0.12)",
        soft: "0 4px 16px rgba(15, 95, 58, 0.08)",
      },
      transitionTimingFunction: {
        "smooth": "cubic-bezier(0.4, 0, 0.2, 1)",
      },
    },
  },
  plugins: [],
} satisfies Config;
