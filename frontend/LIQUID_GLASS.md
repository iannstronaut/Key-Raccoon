# Liquid Glass Design System

Dokumentasi untuk efek glassmorphism (liquid glass) yang diterapkan pada KeyRaccoon Frontend.

## Overview

Liquid Glass adalah teknik desain modern yang menggunakan backdrop blur, transparansi, dan border halus untuk menciptakan efek kaca yang elegan dan futuristik. Efek ini memberikan kesan depth dan premium pada UI.

## Glass Classes

### Base Glass Classes

#### `.glass`
Glass effect standar untuk elemen umum.

```css
background: rgba(16, 17, 17, 0.7);
backdrop-filter: blur(12px);
border: 1px solid rgba(255, 255, 255, 0.08);
```

**Penggunaan:**
- Card standar
- Container biasa
- Background section

#### `.glass-strong`
Glass effect yang lebih kuat dengan opacity lebih tinggi.

```css
background: rgba(16, 17, 17, 0.85);
backdrop-filter: blur(16px);
border: 1px solid rgba(255, 255, 255, 0.1);
```

**Penggunaan:**
- Sidebar
- Modal/Dialog
- Header/Navigation
- Card elevated

#### `.glass-subtle`
Glass effect yang lebih halus dan transparan.

```css
background: rgba(16, 17, 17, 0.5);
backdrop-filter: blur(8px);
border: 1px solid rgba(255, 255, 255, 0.05);
```

**Penggunaan:**
- Hover states
- Nested elements
- Subtle backgrounds

#### `.glass-overlay`
Glass effect untuk overlay/backdrop.

```css
background: rgba(0, 0, 0, 0.6);
backdrop-filter: blur(8px);
```

**Penggunaan:**
- Modal backdrop
- Dropdown overlay
- Sidebar overlay (mobile)

## Component Styles

### Cards

#### `.card`
Card dengan glass effect standar.

**Features:**
- Glass background dengan blur
- Multi-layer shadow untuk depth
- Inset highlight untuk dimensi

**Contoh:**
```tsx
<div className="card p-4">
  <h3>Card Title</h3>
  <p>Card content</p>
</div>
```

#### `.card-elevated`
Card dengan glass effect yang lebih kuat dan shadow lebih dalam.

**Features:**
- Glass-strong background
- Double-ring shadow system
- Enhanced depth dengan multiple shadows
- Inset highlight lebih terang

**Contoh:**
```tsx
<div className="card-elevated p-4">
  <h3>Elevated Card</h3>
  <p>Premium content</p>
</div>
```

### Buttons

#### `.btn-primary`
Primary button dengan glass gradient effect.

**Features:**
- White gradient background
- Soft shadow dengan glow
- Inset highlight
- Hover: lift effect dengan transform

**Contoh:**
```tsx
<button className="btn-primary">
  Create User
</button>
```

#### `.btn-secondary`
Secondary button dengan glass background.

**Features:**
- Transparent glass background
- Backdrop blur
- Subtle border
- Hover: brightness increase

**Contoh:**
```tsx
<button className="btn-secondary">
  Cancel
</button>
```

#### `.btn-ghost`
Ghost button dengan minimal glass effect.

**Features:**
- Very subtle glass background
- Light blur
- Inset highlight
- Hover: background intensifies

**Contoh:**
```tsx
<button className="btn-ghost">
  Learn More
</button>
```

### Inputs

#### `.input-dark`
Input field dengan glass effect.

**Features:**
- Dark glass background
- Backdrop blur
- Inset shadow untuk depth
- Focus: blue glow dengan enhanced glass

**Contoh:**
```tsx
<input 
  type="text" 
  className="input-dark" 
  placeholder="Enter text..."
/>
```

### Badges

#### `.badge`
Base badge dengan glass effect.

**Features:**
- Glass background dengan blur
- Subtle border
- Inset highlight

#### `.badge-success`
Success badge dengan green glass glow.

**Features:**
- Green tinted glass
- Green border
- Soft green glow shadow

#### `.badge-danger`
Danger badge dengan red glass glow.

**Features:**
- Red tinted glass
- Red border
- Soft red glow shadow

#### `.badge-warning`
Warning badge dengan yellow glass glow.

**Features:**
- Yellow tinted glass
- Yellow border
- Soft yellow glow shadow

**Contoh:**
```tsx
<span className="badge badge-success">Active</span>
<span className="badge badge-danger">Error</span>
<span className="badge badge-warning">Warning</span>
```

## Implementation Examples

### Modal dengan Glass Effect

```tsx
{modalOpen && (
  <div className="fixed inset-0 z-50 flex items-center justify-center p-4 glass-overlay">
    <div className="glass-strong w-full max-w-md p-5 relative rounded-xl shadow-2xl border border-white/[0.1]">
      <h3>Modal Title</h3>
      <p>Modal content</p>
    </div>
  </div>
)}
```

### Sidebar dengan Glass Effect

```tsx
<aside className="glass-strong border-r border-white/[0.08]">
  <nav>
    <a className="hover:bg-white/[0.05] hover:backdrop-blur-sm">
      Dashboard
    </a>
  </nav>
</aside>
```

### Card Grid dengan Glass Effect

```tsx
<div className="grid grid-cols-3 gap-4">
  <div className="card-elevated p-4 hover:shadow-lg transition-all">
    <h4>Stat 1</h4>
    <p className="text-2xl">100</p>
  </div>
  <div className="card-elevated p-4 hover:shadow-lg transition-all">
    <h4>Stat 2</h4>
    <p className="text-2xl">200</p>
  </div>
</div>
```

## Design Principles

### 1. Layering
Glass effect bekerja paling baik dengan layering yang jelas:
- Background layer (darkest)
- Glass layer (semi-transparent dengan blur)
- Content layer (foreground)

### 2. Contrast
Pastikan kontras yang cukup antara glass element dan background:
- Gunakan border subtle untuk definisi
- Tambahkan shadow untuk depth
- Inset highlight untuk dimensi

### 3. Blur Intensity
Sesuaikan blur berdasarkan hierarchy:
- **4-8px**: Subtle effects, hover states
- **12px**: Standard glass effect
- **16px**: Strong glass, important elements

### 4. Transparency
Balance antara transparency dan readability:
- **0.5**: Very transparent, subtle
- **0.7**: Standard transparency
- **0.85**: Strong, readable

### 5. Shadows
Multi-layer shadows untuk realistic depth:
- Outer shadow: containment
- Inset shadow: dimension
- Glow shadow: accent (untuk badges)

## Browser Support

Glass effects menggunakan `backdrop-filter` yang didukung oleh:
- ✅ Chrome 76+
- ✅ Edge 79+
- ✅ Safari 9+
- ✅ Firefox 103+

**Fallback:**
Untuk browser yang tidak support, background akan tetap terlihat dengan opacity tanpa blur effect.

## Performance Tips

1. **Limit Blur Usage**: Backdrop blur adalah operasi yang expensive. Gunakan dengan bijak.
2. **Fixed Elements**: Hindari blur pada elemen yang sering bergerak/animate.
3. **Mobile**: Pertimbangkan mengurangi blur intensity di mobile untuk performa.
4. **Layer Count**: Batasi jumlah glass layers yang overlap.

## Accessibility

1. **Contrast**: Pastikan text contrast ratio minimal 4.5:1
2. **Focus States**: Glass elements harus memiliki focus indicator yang jelas
3. **Transparency**: Jangan gunakan transparency yang membuat text sulit dibaca

## Examples in KeyRaccoon

### Dashboard Stats
```tsx
<div className="card-elevated p-4 hover:shadow-lg transition-all">
  <p className="text-xs text-muted">Total Users</p>
  <p className="text-2xl font-semibold">1,234</p>
</div>
```

### Navigation Item
```tsx
<NavLink 
  className={({ isActive }) => 
    `px-3 py-2 rounded-lg transition-all ${
      isActive 
        ? 'bg-white/[0.08] backdrop-blur-sm border border-white/[0.1]' 
        : 'hover:bg-white/[0.05] hover:backdrop-blur-sm'
    }`
  }
>
  Dashboard
</NavLink>
```

### System Health Card
```tsx
<div className="card-elevated p-4">
  <h3>System Health</h3>
  <div className="glass-subtle p-3 rounded-lg hover:bg-white/[0.04]">
    <span>Database</span>
    <span className="badge badge-success">Connected</span>
  </div>
</div>
```

## Customization

Untuk menyesuaikan glass effect, edit nilai di `index.css`:

```css
.glass {
  background: rgba(16, 17, 17, 0.7);  /* Adjust opacity */
  backdrop-filter: blur(12px);         /* Adjust blur */
  border: 1px solid rgba(255, 255, 255, 0.08); /* Adjust border */
}
```

## Best Practices

✅ **DO:**
- Gunakan glass effect untuk elevated elements
- Combine dengan shadow untuk depth
- Maintain consistent blur intensity
- Test readability dengan berbagai backgrounds

❌ **DON'T:**
- Overuse glass effect di semua elements
- Gunakan blur terlalu kuat (>20px)
- Sacrifice readability untuk aesthetic
- Animate blur (performance issue)

## Resources

- [CSS Backdrop Filter](https://developer.mozilla.org/en-US/docs/Web/CSS/backdrop-filter)
- [Glassmorphism Design](https://uxdesign.cc/glassmorphism-in-user-interfaces-1f39bb1308c9)
- [Can I Use: Backdrop Filter](https://caniuse.com/css-backdrop-filter)
