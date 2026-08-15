# 🏳️‍🌈 LGBTScript v8.0

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Version](https://img.shields.io/badge/version-8.0-green)](https://github.com/yourusername/lgbtscript)
[![Pride](https://img.shields.io/badge/🏳️‍🌈-LGBTQ+-ff69b4)](https://github.com/yourusername/lgbtscript)

**LGBTScript** is a programming language and IDE built with the LGBTQ+ community at its heart. It combines a powerful, expressive syntax with built-in tools for safety, creativity, and activism. Write code that celebrates diversity.

## 🌟 Features

- **Inclusive Syntax:** Keywords like `lesbian`, `gay`, `queer`, `nonbinary`, `gender`, `pride`, and `rainbow` make the language feel like home.
- **Hate Speech Filter:** A multi-language filter with morphological analysis protects users and automatically detects and blocks hate speech in output.
- **Built-in LGBTQ+ Functions:** Create coming-out letters, generate pride flags, find safe spaces, and more — all from your code.
- **AI Image Generation:** Use `lgbtImg(prompt, width, height)` to generate pride-themed images via Pollinations.ai.
- **Encrypted Journal:** Keep a secure diary with `lgbtJournal` and AES-256 encryption.
- **Full IDE:** The LGBTScript IDE features syntax highlighting, integrated console, one-click compilation to `.exe`, and a pride-themed UI.
- **Standalone Compilation:** Convert your `.rainbow` scripts into Windows executables with no external dependencies.

## 📚 Documentation

Full documentation is available in the [docs](docs/) folder or at [lgbtscript.dev](https://lgbtscript.dev).

### Quick Start

1. **Installation:** Download the latest release from the [Releases](https://github.com/yourusername/lgbtscript/releases) page. Place `rb.exe` and `rb_ide.exe` in a folder.
2. **Run a Script:** `rb.exe -lgbt my_script.rainbow`
3. **Open the IDE:** `rb_ide.exe` to start the visual development environment.

### Basic Syntax

```go
@ This is a comment

// Variables
lesbian name = "Alex";
gay age = 25;
queer pi = 3.14159;
nonbinary isPride = true;
gender fruits = ["apple", "banana", "cherry"];

// Constants
asexual GRAVITY = 9.81;

// Functions
rainbow greet(name) {
    comingout "Hello, " + name + "!";
    return "Greeted " + name;
}

// Conditionals
cis (age >= 18) {
    comingout "Adult";
} nocis (age > 12) {
    comingout "Teen";
} cis {
    comingout "Child";
}

// Loops
pride (age > 0) {
    comingout "Countdown: " + age;
    age = age - 1;
}

sex (gay i = 0; i < 5; i = i + 1) {
    comingout "i = " + i;
}

// Error Handling
try {
    comingout "Trying risky operation...";
} catch {
    comingout "An error occurred!";
}
🛠️ Built-in Functions
LGBTScript includes a rich set of built-in functions for support, creativity, and safety:

Function	Description
msg(title, text, [type], [icon])	Windows MessageBox. Returns button pressed.
lgbtImg(prompt, width, height)	Generates an AI image.
comingOutLetter(name, recipient, relationship, [style])	Creates a coming-out letter.
transSupport(query)	Returns trans-specific support info.
findAllies(place, radius)	Lists nearby LGBTQ+ friendly spaces.
lgbtHistory([year], [country])	Lists historical events.
lgbtJournal(action, text, [mood])	Encrypted journaling.
createFlag(type, [size])	Renders a pride flag in the console.
checkPrivilege(country)	Rates safety for LGBTQ+ people.
hateMap([region])	Displays hate levels per region.
findMentor([interest], [experience])	Recommends a mentor.
See the full list for all 30+ functions.

🛡️ Hate Speech Filter
The filter is automatically applied to all comingout and msg calls. It supports 20+ languages with morphological variants.

go
lesbian userText = "Some offensive phrase...";
if (checkHate(userText).has_hate) {
    comingout filterHate(userText);  // Replaces slurs with 🌈[blocked]🌈
} else {
    comingout userText;
}
You can also:

enableHateFilter() / disableHateFilter()

checkHate(text) – returns a JSON object with detection details.

filterHate(text) – returns a filtered string.

📁 Directives
#inclusive "file.rainbow" – include another script. Paths are sandboxed.

💻 LGBTScript IDE
The IDE provides a VS Code-like experience with:

Dark theme with pride accents

Syntax highlighting for all keywords and types

Integrated console with error navigation

One‑click execution and compilation

AI image generation menu

Rainbow‑themed toolbar and status bar

Shortcuts:

Key	Action
F5	Run script
Ctrl+O	Open file
Ctrl+S	Save file
Ctrl+G	Generate image
📄 License
This project is licensed under the Apache License, Version 2.0 – see the LICENSE file for details.

text
Copyright 2024 LGBTScript Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
🤝 Contributing
We welcome contributions from everyone. Please read our Contributing Guidelines and Code of Conduct.

💬 Community
Discord Server

Twitter / X

Matrix Space

🌈 Code with Pride. Code for Everyone.
