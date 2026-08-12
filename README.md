# 🏳️‍🌈 LGBTScript

**LGBTScript** is an inclusive programming language designed to support and celebrate the LGBTQ+ community. It combines an easy-to-learn, affirming syntax with powerful features for web development, server applications, and social projects.

> ⚠️ **IMPORTANT LICENSE NOTICE**  
> This project is distributed **WITHOUT A LICENSE**. All rights are reserved.  
> **Commercial use, redistribution, modification, or any form of reuse without explicit written permission from the author is strictly prohibited.**

---

## 📋 Table of Contents

- [Philosophy](#philosophy)
- [Code of Conduct & Prohibited Language](#code-of-conduct--prohibited-language)
- [Features](#features)
- [Data Types](#data-types)
- [Keywords](#keywords)
- [Syntax Examples](#syntax-examples)
- [Complete Function Reference](#complete-function-reference)
- [Installation & Usage](#installation--usage)
- [Building from Source](#building-from-source)
- [License](#license)

---

## 🌈 Philosophy

LGBTScript was created with a simple yet powerful belief:

- **Technology should be inclusive** — everyone deserves to see themselves represented in the tools they use.
- **Programming should be joyful** — using affirming language makes coding more accessible and welcoming.
- **Community matters** — built-in social functions connect developers with resources, safe spaces, and support.
- **Respect is non-negotiable** — this language is a safe space for LGBTQ+ individuals and allies alike.

---

## 🚫 Code of Conduct & Prohibited Language

**This language is designed to be a safe, inclusive space.**

The following are **strictly prohibited** in any LGBTScript code:

- ❌ Slurs or derogatory terms targeting LGBTQ+ individuals or any marginalized group
- ❌ Hate speech of any kind, in code, comments, or variable names
- ❌ Offensive terminology that demeans, ridicules, or discriminates
- ❌ Any language that violates the spirit of inclusivity and respect

> Using such language will result in **syntax errors** and is a violation of the project's core values. This is not just a technical restriction — it is a moral commitment.

---

## ✨ Features

- 🏳️‍🌈 **Inclusive Keywords** — Programming keywords that reflect LGBTQ+ support and affirmation
- 🚀 **High Performance** — Written in Go for fast execution and low memory footprint
- 🌐 **Web Development** — Built-in HTTP server and routing capabilities
- 🤝 **Social Functions** — LGBTQ+ organization support, safe space lookup, crisis help
- 📦 **Self-Contained** — Single executable, no external dependencies
- 🔧 **Compilation** — Compile scripts to standalone `.exe` files
- 💻 **Cross-Platform** — Works on Windows, Linux, and macOS
- 🛡️ **Safe & Inclusive** — Built with respect and community values at its core

---

## 📊 Data Types

| Type | Keyword | Description | Default |
|------|---------|-------------|---------|
| String | `lesbian` | Text data | `""` |
| Integer | `gay` | Whole numbers | `0` |
| Float | `trans` | Decimal numbers | `0.0` |
| Boolean | `nonbinary` | true/false | `false` |
| Array | `gender` | List of values | `[]` |
| Constant | `asexual` | Immutable value | — |

---

## 🔑 Keywords

| Keyword | Purpose |
|---------|---------|
| `comingout` | Print to console |
| `cis` | Conditional (if) |
| `nocis` | Else / else if |
| `pride` | While loop |
| `sex` | For loop |
| `rainbow` | Function declaration |
| `return` | Return from function |
| `try` / `catch` | Error handling |
| `queer` | Class declaration |
| `extends` | Class inheritance |
| `new` | Create object |
| `this` / `super` | Access current/parent object |

---

## 📝 Syntax Examples

### Hello World
```rainbow
RAINBOW main() {
    comingout "🏳️‍🌈 Hello, world from LGBTScript!";
}

main();
Variables
rainbow
lesbian name = "Alex";
gay age = 25;
trans pi = 3.14159;
nonbinary isActive = true;
gender colors = ["red", "green", "blue"];
asexual PI = 3.14159;
Control Flow
rainbow
// If statement
cis (age >= 18) {
    comingout "You are an adult";
} nocis {
    comingout "You are a minor";
}

// While loop
pride (counter < 10) {
    comingout "Counter: " + counter;
    counter++;
}

// For loop
sex (gay i = 0; i < 5; i++) {
    comingout "i = " + i;
}
Functions
rainbow
RAINBOW greet(name) {
    comingout "Hello, " + name;
    return "Greeting sent";
}

greet("Alex");
📚 Complete Function Reference
🌐 Server Functions
Function	Description
createServer(name, port)	Create an HTTP server
addRoute(server, method, path, handler)	Add a route and handler
startServer(name)	Start the server
getServerStatus(name)	Check server status
stopServer(name)	Stop the server
🤝 Social Functions
Function	Description
getLGBTResources()	Get LGBTQ+ organization data
findSafeSpace(place, city)	Find safe spaces for LGBTQ+ individuals
getCrisisSupport(region)	Get crisis support information by region
getDailyAffirmation()	Receive a positive daily affirmation
defineTerm(term)	Define LGBTQ+ related terminology
getComingOutTips(audience)	Get coming out tips for a specific audience
getPrideParadeInfo(city)	Get information about Pride parades in a city
getLGBTQFriendlyCities()	Get a list of LGBTQ+ friendly cities worldwide
📁 File System Functions
Function	Description
readFile(filename)	Read a file's contents
writeFile(filename, content)	Write content to a file
fileExists(filename)	Check if a file exists
🔤 String & Array Functions
Function	Description
split(text, delimiter)	Split a string by delimiter
replace(text, old, new)	Replace substring in text
length(value)	Get length of string or array
🛠 Utility Functions
Function	Description
random(min, max)	Generate a random number
sleep(ms)	Sleep for milliseconds
getTime()	Get current timestamp
🌍 HTTP & JSON Functions
Function	Description
httpGet(url)	Perform an HTTP GET request
httpPost(url, data)	Perform an HTTP POST request
jsonParse(json)	Parse JSON string to object
🔐 Cryptographic Functions
Function	Description
md5(text)	Generate MD5 hash
sha256(text)	Generate SHA256 hash
🚀 Installation & Usage
Download
Download the latest rb.exe (Windows) or rb (Linux/macOS) from the Releases page.

Run a Script
bash
# Run .rainbow file
rb.exe script.rainbow

# Run from command line
rb.exe -c 'comingout "Hello";'

# Show tokens (lexer output)
rb.exe -tokens script.rainbow

# Show AST (Abstract Syntax Tree)
rb.exe -ast script.rainbow

# Compile to standalone .exe
rb.exe -b script.rainbow app.exe
Quick Start Example
bash
# Create a script
echo 'RAINBOW main() { comingout "🌈 Hello!"; } main();' > hello.rainbow

# Run it
rb.exe hello.rainbow
🏗️ Building from Source
bash
# Clone repository
git clone https://github.com/transeva/LGBTScript
cd LGBTScript

# Build (requires Go installed)
go build -o rb rb.go
Prerequisites
Go 1.16 or higher

📄 License
This project is distributed WITHOUT A LICENSE.

All rights are reserved by the author (transeva). This means:

❌ You may not use this software for commercial purposes.

❌ You may not redistribute this software.

❌ You may not modify or create derivative works.

❌ You may not use this software in any way without explicit written permission.

If you wish to use LGBTScript in any capacity beyond personal, non-commercial exploration, you must contact the author for explicit permission.

💖 Support & Community
LGBTScript is a passion project built with love and pride. If you appreciate this work and want to support its development, please reach out to the author.

🐛 Report issues: GitHub Issues

💬 Contact: Reach out to the author directly via GitHub

🌟 Acknowledgments
The LGBTQ+ community for the inspiration and the ongoing fight for visibility and equality

The Go programming language for making this project possible

Everyone who believes that technology can and should be inclusive

🏳️‍🌈 Made with love, pride, and a belief in a better, more inclusive world.

text

---

Both documents are now complete with:
- Full documentation of all features and functions
- License warning (NO LICENSE - all rights reserved)
- Philosophy section explaining the purpose
- Clear Code of Conduct with prohibited language policy
- Complete function reference with all available functions
