# Quick Start Guide

## 🚀 Get Started in 3 Steps

### 1. Install
```bash
curl -fsSL https://raw.githubusercontent.com/dh85/outfitpicker/main/install.sh | bash
```

### 2. Organize Your Photos
Create folders for your outfit types:
```
My Outfits/
├── Work/
├── Casual/
├── Date Night/
└── Gym/
```

### 3. Run It
```bash
outfitpicker
```
Follow the setup wizard, then pick outfits!

## 📱 Daily Usage

| Command | What it does |
|---------|-------------|
| `outfitpicker` | Open the main menu |
| `outfitpicker --quick` | Get instant random outfit |
| `outfitpicker --quick --category Work` | Get instant work outfit |

## 🎯 Menu Options

| Key | Action |
|-----|--------|
| `1,2,3...` | Choose specific outfit folder |
| `r` | Random outfit from all folders |
| `s` | Show what you've already picked |
| `u` | Show what you haven't picked |
| `q` | Quit |

## 🔧 Quick Fixes

| Problem | Solution |
|---------|----------|
| "No outfits available" | `outfitpicker-admin cache clear --all` |
| Changed outfit folder location | `outfitpicker --set-root "/new/path"` |
| Want to start fresh | `outfitpicker config reset` |

## 💡 Pro Tips

- **File names matter**: Use descriptive names like `blue-jeans-white-tee.jpg`
- **Trust the randomness**: Let it surprise you!
- **Reset seasonally**: Clear your history when seasons change
- **Mix loose files**: Photos outside folders work too!

---
📖 **Need more help?** See the full [User Guide](USER_GUIDE.md)