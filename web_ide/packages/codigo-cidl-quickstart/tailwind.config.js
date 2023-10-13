/** @type {import('tailwindcss').Config} */
module.exports = {
  mode: 'jit',
  content: ['./src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        lightGray: '#3e3e42',
        darkGray: '#252526',
        deepDarkGray: '#1e1e1e'
      }
    }
  },
  plugins: [require('@tailwindcss/typography')]
};
