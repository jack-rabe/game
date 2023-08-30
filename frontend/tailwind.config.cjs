/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.svelte', './src/app.html'],
	theme: {
		extend: {}
	},
	plugins: [require('daisyui')],
	daisyui: {
		themes: ['light', 'dark', 'cyberpunk']
	}
};
