# 🏳️‍🌈 LGBTScript

[![LGBTQ+ Friendly](https://img.shields.io/badge/LGBTQ+-Friendly-ff69b4.svg)](https://github.com)
[![License: UNLICENSED](https://img.shields.io/badge/License-UNLICENSED-red.svg)](https://opensource.org/licenses/unlicense)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)](https://golang.org)

**LGBTScript** is an inclusive programming language designed to support the LGBTQ+ community. It combines an easy-to-learn syntax with powerful features for web development, server applications, and social projects.

> ⚠️ **IMPORTANT: This project is distributed WITHOUT A LICENSE. All rights reserved. Commercial use or redistribution without explicit permission is prohibited.**

## 🌈 Features

- 🏳️‍🌈 **Inclusive Keywords** - Programming keywords that reflect LGBTQ+ support
- 🚀 **High Performance** - Written in Go for fast execution
- 🌐 **Web Development** - Built-in HTTP server and routing
- 🤝 **Social Functions** - LGBTQ+ organization support, safe spaces, crisis help
- 📦 **Self-Contained** - Single executable, no dependencies
- 🔧 **Compilation** - Compile scripts to standalone .exe files
- 💻 **Cross-Platform** - Works on Windows, Linux, macOS

## 📊 Data Types

| Type | Keyword | Description | Default |
|------|---------|-------------|---------|
| String | `lesbian` | Text data | `""` |
| Integer | `gay` | Whole numbers | `0` |
| Float | `trans` | Decimal numbers | `0.0` |
| Boolean | `nonbinary` | true/false | `false` |
| Array | `gender` | List of values | `[]` |
| Constant | `asexual` | Immutable value | - |

## 📝 Syntax Overview

### Keywords

| Keyword | Purpose |
|---------|---------|
| `comingout` | Print to console |
| `cis` | Conditional (if) |
| `nocis` | Else/else if |
| `pride` | While loop |
| `sex` | For loop |
| `rainbow` | Function declaration |
| `return` | Return from function |
| `try`/`catch` | Error handling |
| `queer` | Class declaration |
| `extends` | Class inheritance |
| `new` | Create object |
| `this`/`super` | Access current/parent object |

### Example Code

```rainbow
// Hello World
RAINBOW main() {
    COMINGOUT "🏳️‍🌈 Hello, world from LGBTScript!";
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
    COMINGOUT "You are an adult";
} nocis {
    COMINGOUT "You are a minor";
}

// While loop
pride (counter < 10) {
    COMINGOUT "Counter: " + counter;
    counter++;
}

// For loop
sex (gay i = 0; i < 5; i++) {
    COMINGOUT "i = " + i;
}
Functions
rainbow
RAINBOW greet(name) {
    COMINGOUT "Hello, " + name;
    return "Greeting sent";
}

greet("Alex");
🌐 Server Features
LGBTScript includes a built-in HTTP server:

rainbow
// Create server
createServer("myServer", 8080);

// Add route
addRoute("myServer", "GET", "/api/users", handler);

// Start server
startServer("myServer");

// Check status
getServerStatus("myServer");

// Stop server
stopServer("myServer");
Social Functions
Function	Description
getLGBTResources()	Get LGBTQ+ organization data
findSafeSpace(place, city)	Find safe spaces
getCrisisSupport(region)	Crisis support information
getDailyAffirmation()	Daily affirmation
defineTerm(term)	Define a term
getComingOutTips(audience)	Coming out tips
getPrideParadeInfo(city)	Pride parade information
getLGBTQFriendlyCities()	LGBTQ+ friendly cities
Built-in Functions
Function	Description
readFile(filename)	Read file
writeFile(filename, content)	Write to file
fileExists(filename)	Check file existence
split(text, delimiter)	Split string
replace(text, old, new)	Replace substring
length(value)	String/array length
random(min, max)	Random number
sleep(ms)	Sleep in milliseconds
getTime()	Current time
httpGet(url)	HTTP GET request
httpPost(url, data)	HTTP POST request
jsonParse(json)	Parse JSON
md5(text)	MD5 hash
sha256(text)	SHA256 hash
🚀 Installation & Usage
Download
Download the latest rb.exe (Windows) or rb (Linux/macOS) from the releases page.

Run Script
bash
# Run .rainbow file
rb.exe script.rainbow

# Run from command line
rb.exe -c 'comingout "Hello";'

# Show tokens
rb.exe -tokens script.rainbow

# Show AST
rb.exe -ast script.rainbow

# Compile to .exe
rb.exe -b script.rainbow app.exe
Example
bash
# Create a script
echo 'RAINBOW main() { COMINGOUT "🌈 Hello!"; } main();' > hello.rainbow

# Run it
rb.exe hello.rainbow
🏗️ Building from Source
bash
# Clone repository
git clone https://github
