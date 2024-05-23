/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      fontSize: {
        xxs: "0.625rem",
        xxxs: "0.5rem",
      },
    },
  },
  plugins: [],
}

