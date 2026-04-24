/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Primary
        'bg-deep': '#07080a',
        'surface-100': '#101111',
        'surface-card': '#1b1c1e',
        'button-fg': '#18191a',
        // Raycast accents
        'raycast-red': '#FF6363',
        'raycast-blue': '#55b3ff',
        'raycast-green': '#5fc992',
        'raycast-yellow': '#ffbc33',
        // Neutrals & Text
        'text-primary': '#f9f9f9',
        'text-secondary': '#cecece',
        'text-tertiary': '#c0c0c0',
        'text-muted': '#9c9c9d',
        'text-dim': '#6a6b6c',
        'text-dark': '#434345',
        // Borders
        'border-subtle': 'rgba(255, 255, 255, 0.06)',
        'border-light': 'rgba(255, 255, 255, 0.08)',
        'border-medium': '#252829',
        'border-dark': '#2f3031',
        // Glows
        'blue-transparent': 'hsla(202, 100%, 67%, 0.15)',
        'red-transparent': 'hsla(0, 100%, 69%, 0.15)',
        'warm-glow': 'rgba(215, 201, 175, 0.05)',
      },
      fontFamily: {
        inter: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['GeistMono', 'ui-monospace', 'SFMono-Regular', 'Roboto Mono', 'Menlo', 'Monaco', 'monospace'],
      },
      letterSpacing: {
        'body': '0.2px',
        'button': '0.3px',
        'tight': '0.1px',
      },
      borderRadius: {
        'pill': '86px',
      },
      boxShadow: {
        'card': 'rgb(27, 28, 30) 0px 0px 0px 1px, rgb(7, 8, 10) 0px 0px 0px 1px inset',
        'button': 'rgba(255, 255, 255, 0.05) 0px 1px 0px 0px inset, rgba(255, 255, 255, 0.25) 0px 0px 0px 1px, rgba(0, 0, 0, 0.2) 0px -1px 0px 0px inset',
        'warm-glow': 'rgba(215, 201, 175, 0.05) 0px 0px 20px 5px',
        'blue-glow': 'rgba(0, 153, 255, 0.15) 0px 0px 20px 5px',
        'red-glow': 'rgba(255, 99, 99, 0.15) 0px 0px 20px 5px',
      },
    },
  },
  plugins: [],
}
