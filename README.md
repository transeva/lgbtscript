# 🏳️‍🌈 LGBTScript – Language for the LGBTQ+ Community

**LGBTScript** is a programming language and integrated development environment built with inclusivity, creativity, and self-expression at its core. It is designed to provide a safe, empowering, and productive environment for developers from the LGBTQ+ community and allies.

---

## 📖 Table of Contents

- [Philosophy](#philosophy)
- [Features](#features)
- [Getting Started](#getting-started)
- [Language Syntax](#language-syntax)
- [Built-in Functions](#built-in-functions)
- [LGBTScript IDE](#lgbtscript-ide)
- [Examples](#examples)
- [License](#license)
- [Contributing](#contributing)

---

## 🌈 Philosophy

LGBTScript was created to establish a programming language and an ecosystem that:

- **Celebrate Diversity** – Keywords and constructs are designed to be inclusive and affirming.
- **Promote Safety** – Built-in hate speech detection and filtering make online spaces safer.
- **Foster Creativity** – Tools for generating art, building chat servers, and sharing resources.
- **Empower Developers** – Provides a full‑featured IDE with syntax highlighting, debugging, and visual tools.

The language combines the familiarity of C‑like syntax with unique keywords and functions that reflect LGBTQ+ experiences and values.

---

## ✨ Features

### Core Language
- **Type System**: `gay` (int), `lesbian` (string), `queer` (float), `nonbinary` (bool), `gender` (array)
- **Control Flow**: `cis` (if/else if/else), `pride` (while), `sex` (for loop)
- **Functions**: `rainbow` (function declaration), `export rainbow` (export)
- **Constants**: `asexual` (constant declaration)
- **Error Handling**: `try` / `catch`
- **Comments**: `@ comment text`
- **Includes**: `#inclusive "file.rainbow"`

### Built-in Functions (selected)
- **Web & Chats**: `createLGBTChat`, `startLGBTChat`, `sendLGBTChatMessage`, `listLGBTChats`
- **Web Services**: `createServer`, `addRoute`, `getServerStatus`
- **Social Good**: `antiHomoPhobe`, `lgbtImg`, `getLGBTResources`, `findSafeSpace`, `getDailyAffirmation`, `defineTerm`
- **Hate Speech Filter**: `checkHate`, `filterHate`, `addHateSlur`, `getHateStats`
- **Utilities**: `msg` (Windows dialog), `sleep`, `random`, `httpGet`, `jsonParse`

### IDE Features
- **Dark+ Visual Studio Code Theme** with LGBTQ+ accent colors
- **Syntax Highlighting** for all LGBTScript keywords
- **Error Marking** with click‑to‑navigate
- **Integrated Log Viewer** with error parsing
- **One‑Click Compilation** to standalone `.exe`
- **Image Generation** (Pollinations.ai integration)
- **Async Highlighting** for large files
- **Toolbar and Keyboard Shortcuts** (F5: run, Ctrl+O: open, etc.)
- **Built‑in Status Bar** showing line/column position

### Safety & Moderation
- **20+ Languages Supported** for hate speech detection
- **Real-time Chat Filtering** with `sendLGBTChatMessage`
- **Automatic Logging** of hate speech events
- **Customizable Slur Lists** (add/remove)
- **Crisis Support Resources** built‑in (`getCrisisSupport`)

---

## 🚀 Getting Started

### Installation
1. Download `rb.exe` (the interpreter) from the releases page.
2. Place `rb.exe` in your project folder or system path.
3. Download `LGBTScriptIDE.exe` (optional) for the full IDE experience.

### Quick Start (Command Line)
```bash
# Run a script
rb.exe -lgbt myscript.rainbow

# Run inline code
rb.exe -c 'comingout "Hello, World!"'

# Generate an image
rb.exe -lgbtimg "LGBTQ+ pride rainbow" output.png

# Compile to .exe
rb.exe -b myscript.rainbow app.exe
Quick Start (IDE)
Launch LGBTScriptIDE.exe

Write code in the editor

Press F5 to run your script

Use Ctrl+S to save, Ctrl+O to open

Click Compile to generate a standalone executable

Click Image to generate LGBTQ+ themed artwork

📝 Language Syntax
Variable Declaration
rainbow
gay age = 25;               // integer
lesbian name = "Alex";      // string
queer height = 175.5;       // float
nonbinary isActive = true;  // boolean
gender colors = ["red", "blue", "green"]; // array
asexual PI = 3.14159;       // constant
Control Flow
rainbow
cis (age >= 18) {
    comingout "Adult";
} nocis (age < 18) {
    comingout "Minor";
} cis {
    comingout "Fallback";
}

pride (i < 10) {
    comingout "Count: " + i;
    i = i + 1;
}

sex (gay i = 0; i < 5; i = i + 1) {
    comingout "For loop: " + i;
}
Functions
rainbow
rainbow greet(name) {
    comingout "Hello, " + name + "!";
    return "Greeted";
}

export rainbow publicFunction() {
    comingout "Exported function";
}
Error Handling
rainbow
try {
    gay x = 10 / 0;
} catch {
    comingout "Error occurred: " + error;
}
Comments & Includes
rainbow
@ This is a comment
#inclusive "lib.rainbow"
🛠️ Built-in Functions Reference
Web & Chat
Function	Description
createLGBTChat(name, port, [maxMessages])	Creates a chat server
startLGBTChat(name, [port])	Starts the chat server
stopLGBTChat(name)	Stops the chat server
sendLGBTChatMessage(chat, user, msg)	Sends a message to a chat
getLGBTChatMessages(chat, [limit])	Retrieves message history
getLGBTChatStats(chat)	Shows chat statistics
listLGBTChats()	Lists all active chats
createServer(name, port)	Creates an HTTP server
addRoute(server, method, path, handler)	Adds a route
getServerStatus(name)	Shows server status
listServers()	Lists all servers
Social & Empowerment
Function	Description
antiHomoPhobe([duration])	Activates physical anti‑homophobe mode
lgbtImg(prompt, width, height)	Generates LGBTQ+ themed image
getLGBTResources([country], [type])	Lists support organizations
findSafeSpace(place, city, [radius])	Finds LGBTQ+ friendly places
getCrisisSupport(region, [type])	Crisis support contacts
getLGBTQLaws(country, [category])	Legal information per country
getDailyAffirmation([theme])	Returns a daily affirmation
moodCheck([moods], [suggest])	Emotional state check
guidedBreathing([minutes], [theme])	Guided breathing exercise
defineTerm(term, [lang])	Defines LGBTQ+ terms
Hate Speech Filter
Function	Description
checkHate(text)	Checks for hate speech
filterHate(text)	Filters hate speech
enableHateFilter()	Enables the filter
disableHateFilter()	Disables the filter
addHateSlur(lang, slur)	Adds a new slur
removeHateSlur(lang, slur)	Removes a slur
getHateLog()	Shows the hate log
clearHateLog()	Clears the log
getHateStats()	Shows filter statistics
Utilities
Function	Description
msg(title, text, [type], [icon])	Displays Windows dialog
sleep(ms)	Pauses execution
random(min, max)	Generates a random number
httpGet(url)	Performs HTTP GET request
jsonParse(json)	Parses JSON string
🎨 LGBTScript IDE
The LGBTScript IDE is a fully‑featured development environment built with PureBasic.

Interface Features
Dark+ Theme with LGBTQ+ accent colors (red, orange, yellow, green, blue, violet, pink)

Syntax Highlighting for keywords, strings, comments, numbers, operators, variables, functions

Line Numbers with error markers

Real‑time Highlighting for large files (async processing)

Integrated Log Viewer with error click navigation

Status Bar with file info, line/column position

Rainbow‑colored Toolbar buttons with hover effects

Splitter between editor and log

Keyboard Shortcuts
Shortcut	Action
F5	Run script
Ctrl+O	Open file
Ctrl+S	Save file
Ctrl+N	New file
Ctrl+G	Generate image
Ctrl+X/C/V	Cut/Copy/Paste
Ctrl+Z/Y	Undo/Redo
Escape	Stop execution
Image Generation
Built‑in integration with Pollinations.ai (free)

Custom prompt, size, and filename

Image preview with "Open Folder" and "View in Paint" options

LGBTQ+ themed prompts with randomized elements

Compilation
Creates standalone .exe files

Embedded script support (self‑extracting executables)

Full error reporting and logging

📚 Examples
Hello World
rainbow
rainbow main() {
    comingout "Hello, World!";
}
main();
Chat Server
rainbow
rainbow main() {
    createLGBTChat("pride", 8080);
    startLGBTChat("pride");
    sendLGBTChatMessage("pride", "Admin", "Welcome to Pride Chat!");
    comingout getLGBTChatMessages("pride");
}
main();
Web Server
rainbow
rainbow handler(query, body) {
    comingout "Received: " + body;
    return "OK";
}

rainbow main() {
    createServer("api", 3000);
    addRoute("api", "POST", "/data", handler);
    startServer("api");
}
main();
Image Generation
rainbow
rainbow main() {
    lgbtImg("rainbow flag, pride celebration, diversity", 1024, 768);
}
main();
Hate Filter Example
rainbow
rainbow main() {
    lesbian text = "This is a test message";
    gay hasHate = checkHate(text);
    comingout "Hate detected: " + hasHate;
    comingout filterHate(text);
}
main();
📄 License
Copyright 2025 LGBTScript Project

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at:

text
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

🤝 Contributing
We welcome contributions from everyone, especially members of the LGBTQ+ community.

How to Contribute
Fork the repository

Create your feature branch (git checkout -b feature/amazing-feature)

Commit your changes (git commit -m 'Add some amazing feature')

Push to the branch (git push origin feature/amazing-feature)

Open a Pull Request

Code of Conduct
Be respectful and inclusive.

Use affirming and inclusive language.

No hate speech, discrimination, or harassment.

All contributions will be reviewed with kindness and respect.

Development Setup
Language: PureBasic (for IDE), Go (interpreter)

Build: rb.exe is compiled from Go source

IDE: PureBasic 6.x or higher

Reporting Issues
Use the GitHub Issue Tracker

Include your OS, version, and steps to reproduce

Provide logs or screenshots if applicable

🌟 Support & Community
GitHub: https://github.com/yourusername/lgbtscript

Discord: Join the community for support and discussion

Website: https://lgbtlang.org

Email: support@lgbtlang.org

Made with love by the LGBTQ+ community, for the LGBTQ+ community.
