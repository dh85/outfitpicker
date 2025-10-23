# Outfit Picker User Guide

Welcome to Outfit Picker! This guide will help you get started with organizing and selecting your outfits effortlessly.

## What is Outfit Picker?

Outfit Picker is a simple tool that helps you:
- Organize your outfits into folders
- Randomly select outfits you haven't worn recently
- Keep track of what you've already picked
- Never run out of fresh outfit ideas

Think of it as your personal stylist that remembers what you've worn and suggests something new each time!

## Getting Started

### Step 1: Install Outfit Picker

**The Easy Way (Recommended):**
```bash
curl -fsSL https://raw.githubusercontent.com/dh85/outfitpicker/main/install.sh | bash
```

**Other Options:**
- **Mac users**: `brew install dh85/tap/outfitpicker`
- **Windows users**: Download from [releases page](https://github.com/dh85/outfitpicker/releases)

### Step 2: Set Up Your Outfit Folder

1. **Create a main folder** for all your outfits (e.g., "My Outfits" on your Desktop)

2. **Organize by categories** - Create subfolders like:
   - `Work Outfits`
   - `Casual`
   - `Date Night`
   - `Gym`
   - `Beach`
   - `Winter`

3. **Add your outfits** - Put outfits in the appropriate folders:
   ```
   My Outfits/
   ├── Work Outfits/
   │   ├── blue-blazer-outfit.jpg
   │   ├── black-dress-suit.jpg
   │   └── casual-friday.jpg
   ├── Date Night/
   │   ├── red-dress.jpg
   │   └── black-jeans-nice-top.jpg
   └── Casual/
       ├── weekend-comfy.jpg
       └── shopping-outfit.jpg
   ```

### Step 3: First Time Setup

1. **Open your terminal or command prompt**
2. **Type**: `outfitpicker`
3. **Follow the setup wizard**:
   - It will ask for the path to your outfit folder
   - Type the full path (e.g., `/Users/YourName/Desktop/My Outfits`)
   - Press Enter

**Example Setup:**
```
Welcome to Outfit Picker! 👗

Let's set up your outfit directory.
Please enter the path to your 'Outfits' folder: /Users/Sarah/Desktop/My Outfits

✅ Great! Your outfit directory is set up.
```

## How to Use Outfit Picker

### Basic Usage

Simply type `outfitpicker` in your terminal and follow the menu:

```
📂 Outfit Folders

  [1] 📂 Work Outfits (3 outfits)
  [2] 📂 Date Night (2 outfits) 
  [3] 📂 Casual (2 outfits)

📋 What would you like to do?
  [r] Pick a random outfit for me
  [s] Show me what I've already picked
  [u] Show me what I haven't picked yet
  [q] Quit

Your choice: 
```

### Menu Options Explained

**[1, 2, 3...]** - **Choose a specific category**
- Select a number to see outfits from that folder
- Great when you know what type of outfit you want

**[r]** - **Pick a random outfit for me**
- Selects a random outfit from ALL your folders
- Prioritizes outfits you haven't worn recently
- Perfect for when you want to be surprised!

**[s]** - **Show me what I've already picked**
- Lists all the outfits you've selected before
- Helps you see your recent choices

**[u]** - **Show me what I haven't picked yet**
- Shows fresh outfit options
- Great for seeing what's still available

**[q]** - **Quit**
- Exits the app

### When You Pick an Outfit

After selecting an outfit, you'll see:

```
🎲 I picked this outfit for you: red-dress.jpg

Do you want to (k)eep it, (s)kip it, or (q)uit? 
```

**Your Options:**
- **[k] Keep it** - "Yes, I'll wear this!" (marks it as worn)
- **[s] Skip it** - "Not today, show me something else"
- **[q] Quit** - Exit the app

### Quick Mode (For Power Users)

If you want instant results without menus:

```bash
# Pick any random outfit instantly
outfitpicker --quick

# Pick from a specific category
outfitpicker --quick --category "Work Outfits"
```

## Advanced Features

### Working with Loose Files

Don't worry if you have some outfit photos that don't fit into categories! Just put them directly in your main outfit folder:

```
My Outfits/
├── Work Outfits/
│   └── suit.jpg
├── random-cute-outfit.jpg    ← Loose file (totally fine!)
└── vacation-look.jpg         ← Another loose file
```

Outfit Picker will find these and include them in your selections.

### Managing Your Selections

**See what you've picked:**
- Choose `[s]` from the main menu
- Review your recent outfit choices

**Reset your history:**
- Use the admin tool: `outfitpicker-admin cache clear --all`
- This clears all your selection history (fresh start!)

**Clear specific categories:**
- `outfitpicker-admin cache clear "Work Outfits"`
- Only clears history for that folder

## Tips for Best Results

### Outfit Organization Tips

1. **Use clear filenames**: `blue-jeans-white-tee.avatar` instead of `avatar1.avatar`
2. **Consistent categories**: Stick to the same folder names you create initially
3. **Regular updates**: Add new outfits as you get them

### Usage Tips

1. **Check regularly**: Use Outfit Picker daily or weekly for best results
2. **Trust the randomness**: Let it surprise you with combinations you might not have thought of
3. **Reset when needed**: Clear your history seasonally or when you get new clothes
4. **Mix it up**: Use both specific categories and random selection

## Troubleshooting

### "No outfits available"
- **Problem**: All your outfits have been selected recently
- **Solution**: Use `outfitpicker-admin cache clear --all` to reset your history

### "Category not found"
- **Problem**: You typed a category name that doesn't exist
- **Solution**: Check your folder names and try again (case doesn't matter)

### "No input provided"
- **Problem**: The app doesn't know where your outfits are
- **Solution**: Run `outfitpicker` and follow the setup wizard again

### Can't find your outfit folder
- **Problem**: You moved or renamed your outfit folder
- **Solution**: Run `outfitpicker --set-root "/new/path/to/outfits"` with the new location

## Getting Help

### Built-in Help
- Type `outfitpicker --help` for command options
- Type `outfitpicker-admin --help` for admin commands

### Configuration
- **See current settings**: `outfitpicker config show`
- **Change outfit folder**: `outfitpicker config set-root "/new/path"`
- **Reset everything**: `outfitpicker config reset`

## Example Workflow

Here's how Sarah uses Outfit Picker every morning:

1. **Opens terminal** and types `outfitpicker`
2. **Sees her categories**: Work Outfits, Casual, Gym
3. **Types `r`** for a random outfit (she likes surprises!)
4. **Gets suggestion**: "blue-blazer-outfit.avatar"
5. **Types `k`** to keep it (she likes the choice)
6. **Gets dressed** and starts her day!

On days when she knows she needs work clothes:
1. **Types `1`** to select Work Outfits category
2. **Types `r`** for random work outfit
3. **Gets work-appropriate suggestion**

## Why Use Outfit Picker?

- **Saves time**: No more staring at your closet wondering what to wear
- **Reduces decision fatigue**: Let the app decide for you
- **Ensures variety**: Automatically rotates through your options
- **Rediscover forgotten outfits**: Find clothes you forgot you had
- **Seasonal flexibility**: Clear history when seasons change

## Privacy & Data

- **Everything stays on your computer**: No data is sent anywhere
- **Simple file storage**: Your selections are saved in a small file on your computer
- **Easy to reset**: Delete your history anytime with admin commands

---

**Need more help?** Visit the [GitHub repository](https://github.com/dh85/outfitpicker) for technical documentation and support.

**Happy outfit picking!** 👗✨
