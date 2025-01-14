/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./internal/web/views/*.templ"],
  theme: {
    extend: {},
  },
  plugins: [require("@tailwindcss/typography"), require("daisyui")],
};
