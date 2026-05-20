/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Linear-inspired dark palette
        bg: {
          DEFAULT: "#0a0a0b",      // app background, near-black
          surface: "#111113",      // cards, panels
          elevated: "#16161a",     // modals, popovers
          hover: "#1c1c21",        // interactive hover states
        },
        border: {
          DEFAULT: "#1f1f24",      // default borders
          strong: "#2a2a31",       // emphasized borders
        },
        text: {
          DEFAULT: "#e5e5e7",      // primary text
          muted: "#8a8a93",        // secondary text
          subtle: "#5b5b64",       // tertiary text (timestamps, hints)
        },
        accent: {
          DEFAULT: "#7c5cff",      // primary action (Linear purple-ish)
          hover: "#8c6dff",
          subtle: "#7c5cff20",     // 20% alpha for backgrounds
        },
        success: {
          DEFAULT: "#3fb950",
          subtle: "#3fb95020",
        },
        warning: {
          DEFAULT: "#d29922",
          subtle: "#d2992220",
        },
        danger: {
          DEFAULT: "#f85149",
          subtle: "#f8514920",
        },
      },
      fontFamily: {
        sans: ['Inter', '-apple-system', 'BlinkMacSystemFont', 'sans-serif'],
        mono: ['JetBrains Mono', 'Menlo', 'Consolas', 'monospace'],
      },
      fontSize: {
        xs: ['0.75rem', { lineHeight: '1rem' }],
        sm: ['0.8125rem', { lineHeight: '1.25rem' }],
        base: ['0.875rem', { lineHeight: '1.375rem' }],
        lg: ['1rem', { lineHeight: '1.5rem' }],
        xl: ['1.125rem', { lineHeight: '1.625rem' }],
        '2xl': ['1.375rem', { lineHeight: '1.75rem' }],
        '3xl': ['1.75rem', { lineHeight: '2rem' }],
      },
      borderRadius: {
        sm: '4px',
        DEFAULT: '6px',
        md: '8px',
        lg: '10px',
        xl: '12px',
      },
      boxShadow: {
        glow: '0 0 0 1px rgba(124, 92, 255, 0.5), 0 0 20px rgba(124, 92, 255, 0.2)',
      },
    },
  },
  plugins: [],
}