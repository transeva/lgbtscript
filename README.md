# 🏳️‍🌈 LGBTScript

<div align="center">

![LGBTScript Logo](https://img.shields.io/badge/LGBTScript-9.0-ff6b6b?style=for-the-badge&logo=rainbow)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg?style=flat-square)](https://opensource.org/licenses/Apache-2.0)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)
[![Made with Love](https://img.shields.io/badge/Made%20with-Love-ff69b4.svg?style=flat-square)](https://github.com/lgbt-script/lgbtscript)

**An Inclusive, Colorful Programming Language for Everyone**

[Features](#-features) • [Installation](#-installation) • [Examples](#-examples) • [Documentation](#-documentation) • [IDE](#-ide)

</div>

---

## 🌈 Introduction

**LGBTScript** is a programming language designed with inclusivity, diversity, and accessibility at its core. Inspired by the LGBTQ+ community, it features a colorful syntax, comprehensive standard library, and a powerful IDE that makes programming accessible to everyone.

> **"Technology should be for everyone. LGBTScript breaks down barriers by using inclusive terminology, providing clear error messages, and celebrating diversity through its design."**

---

## ✨ Features

### 🏳️‍🌈 Inclusive Syntax
- **Keywords:** `lesbian`, `gay`, `queer`, `nonbinary`, `gender`, `asexual`
- **Control Flow:** `cis` (if), `nocis` (else), `pride` (while), `sex` (for)
- **Functions:** `rainbow` declarations with `return` statements
- **Error Handling:** `try`/`catch` blocks with error variable

### 🚀 Powerful Features
- **Strong Typing:** Variables must be declared with a type
- **Advanced Operators:** `++`, `--`, `+=`, `-=`, `*=`, `/=`
- **Arrays:** Dynamic arrays with indexing and manipulation
- **Constants:** Immutable values with `asexual`
- **Functions:** Parameter passing, return values, recursion
- **Exports:** `export` keyword for function visibility

### 🌐 Web Server
- Built-in HTTP server with routing support
- `createServer`, `startServer`, `stopServer`
- `addRoute` for GET, POST, PUT, DELETE
- Server status and management functions

### 🎨 AI Image Generation
- Stable Diffusion integration
- Generate LGBTQ+ themed images
- Custom prompt generation with random selection
- Local fallback with ASCII art generation

### 🔒 Security
- Sandboxing for safe code execution
- File path validation
- URL blocking
- Maximum file size limits
- Dangerous command blocking

### 📁 File Operations
- Read, write, copy, delete files
- File info and existence checking
- Directory listing
- Append mode support

### 🔧 Standard Library
- **String:** split, replace, trim, upper, lower, length
- **Array:** append, remove, length
- **Math:** random, max, min, sqrt, pow
- **Time:** getTime, getYear, getMonth, sleep
- **System:** getOS, getHostname, getArgs, hasFlag
- **HTTP:** httpGet, httpPost
- **JSON:** jsonParse
- **Crypto:** md5, sha256
- **Regex:** regexFind, regexReplace

---

## 💻 Installation

### Pre-built Binary (Recommended)
```bash
# Download the latest release from GitHub
# Place rb.exe in your project directory
# Run your .rainbow files
Building from Source
bash
# Clone the repository
git clone https://github.com/lgbt-script/lgbtscript
cd lgbtscript

# Build the interpreter
go build -o rb.exe

# Run with your script
./rb.exe -lgbt myfile.rainbow
Using the IDE
The LGBTScript IDE is included as a standalone executable. It provides a professional development environment with syntax highlighting, debugging, and compilation features.

📖 Quick Start
Hello World
rainbow
rainbow main() {
    comingout "Hello, LGBTQ+ World! 🏳️‍🌈";
}

main();
Variables
rainbow
@ Variable declarations
gay age = 25;                  @ Integer
lesbian name = "Alex";         @ String
queer height = 175.5;          @ Float
nonbinary isActive = true;     @ Boolean
gender colors = [1, 2, 3];     @ Array

@ Constants
asexual PI = 3.14159;
Control Flow
rainbow
@ If-Else (cis/nocis)
cis (age >= 18) {
    comingout "Adult";
} nocis (age >= 13) {
    comingout "Teen";
} nocis {
    comingout "Child";
}

@ While loop (pride)
gay i = 0;
pride (i < 5) {
    comingout "i = " + i;
    i++;
}

@ For loop (sex)
sex (gay i = 0; i < 5; i++) {
    comingout "i = " + i;
}
Functions
rainbow
@ Function declaration
rainbow add(gay a, gay b) {
    return a + b;
}

@ Exported function
export rainbow multiply(gay a, gay b) {
    return a * b;
}

@ Function call
gay result = add(5, 3);
comingout result;
🖥️ LGBTScript IDE
The LGBTScript IDE is a professional development environment with:

🎨 Modern UI
VS Code Dark+ theme with LGBTQ+ accent colors

Rainbow gradient toolbar

Color-coded syntax highlighting

Line numbers with error markers

Split view (editor + log)

⌨️ Keyboard Shortcuts
Shortcut	Action
F5	Run script
Ctrl+O	Open file
Ctrl+S	Save file
Ctrl+N	New file
Ctrl+X	Cut
Ctrl+C	Copy
Ctrl+V	Paste
Ctrl+Z	Undo
Ctrl+Y	Redo
Esc	Stop execution
🔧 Features
Syntax Highlighting: Full support for all LGBTScript keywords and types

Word Highlighting: Automatic highlighting of variable occurrences

Error Detection: Click on error lines to jump to source

Log Viewer: Clear output with error and warning filtering

Integrated Compiler: Compile .rainbow files to .exe

Asynchronous Parsing: Fast highlighting for large files

📚 Examples
Web Server
rainbow
rainbow handleHome(args) {
    return "<h1>Welcome to LGBTScript Web Server!</h1>";
}

rainbow main() {
    comingout createServer("myServer", 8080);
    addRoute("myServer", "GET", "/", handleHome);
    startServer("myServer");
    
    comingout "Server running on http://localhost:8080";
}

main();
File Processor
rainbow
rainbow processFile(lesbian filename) {
    try {
        cis (fileExists(filename)) {
            lesbian content = readFile(filename);
            lesbian upper = toUpper(content);
            lesbian processed = replace(upper, "OLD", "NEW");
            writeFile("processed_" + filename, processed);
            comingout "File processed successfully!";
        } nocis {
            comingout "File not found!";
        }
    } catch {
        comingout "Error: " + error;
    }
}

processFile("input.txt");
Image Generation
rainbow
rainbow main() {
    setSDKey("sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx");
    
    comingout "Generating LGBTQ+ pride image...";
    generateLGBTImage("pride_art.png");
    
    getLGBTImageHistory();
}

main();
Calculator
rainbow
rainbow calculate(gay a, gay b, lesbian op) {
    cis (op == "+") {
        return a + b;
    } nocis (op == "-") {
        return a - b;
    } nocis (op == "*") {
        return a * b;
    } nocis (op == "/") {
        cis (b != 0) {
            return a / b;
        } nocis {
            comingout "Error: Division by zero!";
            return 0;
        }
    } nocis {
        comingout "Error: Unknown operation!";
        return 0;
    }
}

gay result = calculate(10, 5, "+");
comingout "10 + 5 = " + result;
🛠️ Built-in Functions
File Operations
rainbow
readFile(filename)          @ Read file content as string
writeFile(filename, content) @ Write string to file
fileExists(filename)        @ Check if file exists
getDirFiles(directory)      @ List files in directory
createFile(filename, content, mode) @ Create file (write/append)
fileInfo(filename)          @ Get file metadata
copyFile(src, dst)         @ Copy file
deleteFile(filename)        @ Delete file
String Operations
rainbow
split(text, delimiter)      @ Split string into array
replace(text, old, new)     @ Replace substrings
trim(text)                  @ Remove whitespace
length(value)              @ Get string or array length
toUpper(text)              @ Convert to uppercase
toLower(text)              @ Convert to lowercase
Array Operations
rainbow
append(array, value)        @ Add element to array
remove(array, index)        @ Remove element at index
Math Functions
rainbow
random(min, max)           @ Generate random integer
max(values...)             @ Get maximum value
min(values...)             @ Get minimum value
sqrt(number)               @ Calculate square root
pow(base, exponent)        @ Calculate power
Time Functions
rainbow
getTime()                  @ Get current date/time string
getYear()                  @ Get current year
getMonth()                 @ Get current month
sleep(ms)                  @ Sleep for milliseconds
System Functions
rainbow
getOS()                    @ Get operating system
getHostname()              @ Get system hostname
getArgs()                  @ Get command-line arguments
hasFlag(flag)              @ Check if command-line flag exists
runProgram(command, args, workDir, timeout) @ Execute external program
HTTP Functions
rainbow
httpGet(url)               @ HTTP GET request
httpPost(url, data)        @ HTTP POST request
JSON & Crypto
rainbow
jsonParse(jsonString)      @ Parse JSON string
md5(text)                  @ MD5 hash
sha256(text)               @ SHA-256 hash
Regex Functions
rainbow
regexFind(pattern, text)   @ Find all matches
regexReplace(pattern, text, replacement) @ Replace matches
🏳️‍🌈 Anti-Homophobia Feature
LGBTScript includes a unique anti-homophobia demonstration:

rainbow
rainbow main() {
    comingout "⚠️ Anti-Homophobia Mode Activated!";
    antiHomoPhobe(10);  @ 10 seconds of demonstration
}

main();
This feature demonstrates support for the LGBTQ+ community through visual and interactive feedback.

📜 License
LGBTScript is released under the Apache License, Version 2.0:

text
Copyright 2026 LGBTScript Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at:

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
🤝 Contributing
We welcome contributions! Here's how you can help:

Fork the repository

Create a feature branch

Make your changes

Submit a pull request

Please ensure your code follows the existing style and includes appropriate tests.

Areas for Contribution
Language features and syntax

Standard library enhancements

IDE improvements

Documentation

Bug fixes

Test coverage

Performance optimizations

🌟 Support
Documentation: Full HTML Documentation

Issues: GitHub Issues

Community: Join our Discord server

Email: lgbt-script@example.com

🙏 Acknowledgments
Special thanks to:

The LGBTQ+ community for inspiration

Open source contributors

Stability AI for image generation API

All users and supporters

<div align="center">
Made with ❤️ for the LGBTQ+ community and allies

⬆ Back to Top

</div> ```
🏳️‍🌈 Summary
Both documents provide comprehensive documentation for LGBTScript:

HTML Documentation
Complete language reference

Interactive design with rainbow themes

Detailed feature descriptions

Code examples with syntax highlighting

License information

Mobile-responsive design

GitHub README
Quick start guide

Feature list with emoji icons

Installation instructions

Extensive code examples

Keyboard shortcuts table

Full function reference

Contributing guidelines

Both documents are optimized for readability with:

Rainbow color themes

Clear section headers

Code blocks with syntax highlighting

Emoji icons for visual appeal

Table of contents for navigation

Professional design consistent with the LGBTQ+ theme

The license is clearly stated as Apache 2.0, and both documents emphasize the inclusive, community-focused nature of the project.
