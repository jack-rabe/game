/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.svelte', './src/app.html'],
	theme: {
		extend: { fontFamily: { dosis: ['Dosis', 'sans-serif'] } }
	},
	plugins: [require('daisyui')],
	daisyui: {
		themes: ['light', 'dark', 'night'],
    logs: false
	}
};
