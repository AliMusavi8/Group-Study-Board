---
name: Frontend Design Best Practices
description: Guidelines and best practices for developing frontend web applications with Angular, HTML, and Tailwind CSS.
---

# Frontend Design Best Practices (Angular + Tailwind CSS)

When modifying or creating new features in the frontend of the application, follow these best practices for building scalable, maintainable, and highly performant interfaces using Angular and Tailwind CSS.

## 1. Architectural Patterns

### Component Reusability
- **Encapsulate Repetitive Styles**: Avoid "class soup" in your HTML templates by keeping shared UI patterns inside Angular components or use the `@apply` directive for primitive elements (like `.btn` or `.card`). 
- **Start in HTML**: Prototype and style directly in the HTML. Only extract classes when patterns are clearly repeated across multiple components.
- **Standalone Components**: Use Angular's Standalone Components (v15+) module-free architecture where appropriate. They integrate seamlessly with Tailwind components.

### Dynamic State and Styling
- **Angular Signals for State**: Use Angular Signals for computing derived state and dynamic styling logic.
- **`[ngClass]` and `[class.X]` Binding**: Use Angular's built-in class bindings to toggle Tailwind utilities based on component state rather than manually adding/removing classes from the DOM.

## 2. Tailwind CSS Conventions

### Theming & Configuration
- **Use `tailwind.config.js`**: All project-specific colors, fonts, breakpoint overrides, and spacing rules must be defined in the `tailwind.config.js` extending the default theme. Avoid hardcoded hex codes (`border-[#abc]`) when possible, map them to theme colors (`border-accent`).
- **Clean Configuration**: Only maintain styles related to your application in `styles.css` using the `@tailwind base; @tailwind components; @tailwind utilities;` directives. Custom CSS should be minimal and limited to complex animations or specific behaviors Tailwind doesn't natively support.

### Structuring HTML Classes
- Organize your Tailwind class lists consistently. A standard ordering looks like:
  1. Layout (e.g., `block`, `flex`, `grid`, `absolute`)
  2. Spacing / Sizing (e.g., `w-full`, `h-10`, `p-4`, `m-2`)
  3. Box Model (e.g., `border`, `rounded-md`)
  4. Typography (e.g., `text-lg`, `font-bold`, `text-center`)
  5. Colors (e.g., `bg-white`, `text-gray-900`)
  6. States (e.g., `hover:bg-gray-100`, `focus:ring`)
- By following this model, components remain intuitive and readable to other developers or AI agents reviewing the codebase.

## 3. Responsive & Accessible Interfaces

### Mobile-First Design
- Use Tailwind's default mobile-first breakpoints. Apply base utility classes for mobile screens, and constrain or modify for larger screens via prefixes (e.g., `md:`, `lg:`, `xl:`).

### Accessibility (a11y)
- **Contrast Ratios**: Check typography and background colors to ensure they adhere to WCAG standards.
- **Interactive Elements**: Always apply `:focus` and `:focus-visible` states to buttons, inputs, and links using `focus:outline-none focus:ring-2 focus:ring-accent` patterns.
- **Semantics**: Do not misuse CSS for layout structure if semantic HTML exists (use `<nav>`, `<aside>`, `<main>`).

## 4. Performance Optimization

### Tree-Shaking
- Ensure that `tailwind.config.js` properly specifies paths to all HTML and TypeScript files within the `content` array (`['./src/**/*.{html,ts}']`) so Tailwind compiler can purge unused styles effectively during production builds.

## 5. TypeScript Best Practices (Angular Context)

### Strong Typing & Interfaces
- **Define Interfaces**: Always create explicit `interface` or `type` definitions for data models, API responses, and component state.
- **Avoid `any`**: Strictly minimize the use of the `any` type. Use `unknown` or Generic types (`<T>`) when the shape of the data is inherently dynamic.
- **Utility Types**: Leverage built-in utility types like `Partial<T>`, `Readonly<T>`, `Pick<T, K>`, and `Omit<T, K>` to manipulate and extend existing interfaces safely without duplication.

### Angular-Specific TypeScript
- **Strict Mode**: Ensure `strict` mode is enabled in `tsconfig.json` to catch null-dereferencing and implicit `any` errors at compile-time.
- **RxJS Subscriptions**: When subscribing to RxJS Observables within components, always ensure they are properly unsubscribed when the component is destroyed. The preferred modern approach is using the `takeUntilDestroyed()` operator or Angular's `AsyncPipe` in the template to avoid memory leaks.
- **Type Guards**: Use custom type guard functions (`function isMyType(value: any): value is MyType`) when dealing with union types or mapping unknown JSON payloads to strict interfaces.
- **Access Modifiers**: Clearly mark service methods and component properties with `private`, `protected`, or `public`. If a property is only used in the HTML template, it should be `public` (or strictly typed as an Angular Signal).
