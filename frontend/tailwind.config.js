module.exports = {
  content: ["./src/**/*.{html,ts}"],
  theme: {
    extend: {
      colors: {
        canvas:         "#f7f4ee",
        ink:            "#1a2e1f",
        accent:         "#16a34a",
        "accent-hover": "#15803d",
        "accent-light": "#dcfce7",
        muted:          "#64748b",
        danger:         "#dc2626",
        "danger-hover": "#b91c1c",
        border:         "#bbf7d0",
      },
      boxShadow: {
        soft:   "0 4px 24px rgba(0, 0, 0, 0.08)",
        canvas: "0 8px 48px rgba(0, 0, 0, 0.15)",
      }
    }
  },
  plugins: []
};
