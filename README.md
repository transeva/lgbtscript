# 🏳️‍🌈 LGBTScript - An Inclusive Programming Language for Everyone

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Version](https://img.shields.io/badge/Version-9.0-brightgreen.svg)](https://github.com/your-username/lgbtscript)
[![Pride](https://img.shields.io/badge/🏳️‍🌈-Pride-red.svg)](https://github.com/your-username/lgbtscript)

LGBTScript is a modern, expressive programming language designed with inclusivity at its core. It combines powerful programming features with a syntax that celebrates LGBTQ+ identity and promotes community support.

## ✨ Features

### 🏳️‍🌈 LGBTQ+ Centered Design
- **Inclusive Keywords**: `gay`, `lesbian`, `trans`, `nonbinary`, `gender`, `queer`, `pride`
- **Empowering Syntax**: `comingout` for output, `cis/nocis` for conditionals, `rainbow` for functions
- **Community First**: Built-in LGBTQ+ resources, crisis support, and community tools

### 🚀 Full-Featured Language
- **Strong Typing**: Integer (`gay`), String (`lesbian`), Float (`trans`), Boolean (`nonbinary`), Array (`gender`)
- **OOP Support**: `queer` classes with inheritance (`extends`), constructors, methods, fields
- **Control Flow**: `cis`/`nocis` (if/else), `pride` (while), `sex` (for)
- **Error Handling**: `try`/`catch` blocks with `error` variable

### 🌐 Web Development
- **HTTP Server**: Create web servers with `createServer()`, `startServer()`, `stopServer()`
- **Routing**: Add routes with `addRoute()` (GET, POST, PUT, DELETE)
- **Request Handling**: Access `query` and `body` parameters in route handlers

### 🛠️ Built-in Functions
- **File I/O**: `readFile()`, `writeFile()`, `fileExists()`, `getDirFiles()`
- **String Manipulation**: `split()`, `replace()`, `trim()`, `toUpper()`, `toLower()`
- **Array Operations**: `append()`, `remove()`, `length()`
- **Math**: `random()`, `max()`, `min()`, `sqrt()`, `pow()`
- **HTTP Client**: `httpGet()`, `httpPost()`, `jsonParse()`
- **Crypto**: `md5()`, `sha256()`
- **System**: `getTime()`, `getYear()`, `getOS()`, `getArgs()`, `hasFlag()`

### 🏳️‍🌈 LGBTQ+ Support Functions
- **Resources**: `getLGBTResources()` - Verified LGBTQ+ organizations (JSON)
- **Crisis Support**: `getCrisisSupport()`, `findSafeSpace()`, `getLGBTQLaws()`
- **Health & Wellness**: `getHRTInfo()`, `findLGBTDoctor()`, `getTransHealthcare()`
- **Community**: `getLGBTQEvents()`, `createLGBTQGroup()`, `findVolunteerOpportunity()`
- **Education**: `getLGBTQBook()`, `getLGBTQPlaylist()`, `getLGBTQMovies()`, `getLGBTQHistory()`
- **Mental Health**: `getDailyAffirmation()`, `moodCheck()`, `guidedBreathing()`, `getComingOutTips()`
- **Specialized**: `getIntersexResources()`, `getNonbinaryGuide()`, `getAsexualResources()`, `getPolyamoryGuide()`

### 🎨 Creative Tools
- **Image Generation**: `lgbtImg(prompt, width, height)` - Generate LGBTQ+ themed images using Pollinations.ai
- **Rainbow Output**: Built-in rainbow colors and pride-themed aesthetics
- **Visual Tools**: Show generated images in a preview window

### 💻 LGBTScript IDE
- **Beautiful Dark Theme**: VS Code inspired with rainbow accents
- **Syntax Highlighting**: Full support for all LGBTScript keywords
- **Code Intelligence**: Smart word highlighting, variable/function detection
- **Async Highlighting**: Fast performance for large files
- **Error Navigation**: Visual error markers with click-to-jump
- **Run & Debug**: Execute scripts directly from the IDE
- **Compile to EXE**: Create standalone Windows executables
- **Keyboard Shortcuts**: F5 to run, Ctrl+O to open, Ctrl+S to save

## 🚀 Installation

### Pre-built Binary
Download the latest release from the [releases page](https://github.com/your-username/lgbtscript/releases)

### From Source
```bash
# Clone the repository
git clone https://github.com/your-username/lgbtscript.git
cd lgbtscript

# Build with PureBasic
# Open LGBTScript.pb in PureBasic IDE and compile
📖 Quick Start
Hello World
rainbow
rainbow main() {
    comingout "🏳️‍🌈 Hello, world! This is LGBTScript!";
}

main();
Variables and Types
rainbow
@ Variable declarations
gay age = 25;
lesbian name = "Alex";
trans height = 1.85;
nonbinary isValid = true;
gender colors = ["red", "orange", "yellow"];

@ Constants
asexual PI = 3.14159;
asexual PRIDE_FLAG = "🏳️‍🌈";

@ Output
comingout "Hello, " + name + "! Age: " + age;
comingout "Height: " + height;
comingout "Colors: " + colors;
Control Flow
rainbow
@ If/Else (cis/nocis)
cis (age >= 18) {
    comingout "You are an adult";
} nocis (age > 60) {
    comingout "You are a senior";
} nocis {
    comingout "You are a minor";
}

@ While loop (pride)
gay i = 0;
pride (i < 5) {
    comingout "i = " + i;
    i++;
}

@ For loop (sex)
sex (gay j = 0; j < 3; j++) {
    comingout "j = " + j;
}
Functions (rainbow)
rainbow
rainbow greet(name) {
    comingout "Hello, " + name + "!";
    return "Greeting sent";
}

rainbow add(a, b) {
    gay sum = a + b;
    return sum;
}

lesbian result = greet("Alex");
gay total = add(5, 3);
QUEER Classes (OOP)
rainbow
queer Person {
    lesbian name;
    gay age;
    nonbinary isActive;
    
    rainbow init(name, age) {
        this.name = name;
        this.age = age;
        this.isActive = true;
    }
    
    rainbow greet() {
        comingout "Hello, I'm " + this.name;
    }
}

queer Student extends Person {
    lesbian school;
    
    rainbow init(name, age, school) {
        super.init(name, age);
        this.school = school;
    }
    
    rainbow study() {
        comingout this.name + " studies at " + this.school;
    }
}

queer alex = new Student("Alex", 20, "University");
alex.greet();
alex.study();
Web Server
rainbow
rainbow main() {
    @ Create server
    createServer("myServer", 8080);
    
    @ Add routes
    rainbow helloHandler() {
        return "Hello, world!";
    }
    addRoute("myServer", "GET", "/hello", helloHandler);
    
    rainbow jsonHandler() {
        return "{\"status\":\"ok\"}";
    }
    addRoute("myServer", "GET", "/json", jsonHandler);
    
    @ Start server
    startServer("myServer");
}

main();
Image Generation
rainbow
rainbow main() {
    comingout "🖼️ Generating Pride image...";
    lgbtImg("rainbow pride flag, LGBTQ+ community, love", 1024, 768);
}

main();
LGBTQ+ Resource Finder
rainbow
rainbow main() {
    @ Get verified LGBTQ+ organizations
    lesbian data = getLGBTResources("Россия", "network", "", "");
    comingout "📊 Resources in Russia:";
    comingout data;
    
    @ Get crisis support
    lesbian support = getCrisisSupport("Москва", "горячая_линия");
    comingout support;
}

main();
Error Handling
rainbow
rainbow main() {
    try {
        gay result = 10 / 0;
        comingout "Result: " + result;
    } catch {
        comingout "Error: " + error;
        comingout "💪 We learn from our mistakes and grow stronger!";
    }
}

main();
🎮 LGBTScript IDE
The LGBTScript IDE provides a beautiful development environment with:

Features
Rainbow-themed UI with VS Code dark theme

Full syntax highlighting for all keywords

Smart code intelligence with word highlighting

Async processing for large files

Clickable error navigation

Integrated console with output logging

File management (open, save, new)

Run scripts directly from IDE

Compile to EXE for standalone executables

Image generation with preview

Keyboard shortcuts for productivity

Keyboard Shortcuts
Shortcut	Action
F5	Run current script
Ctrl+O	Open file
Ctrl+S	Save file
Ctrl+N	New file
Ctrl+G	Generate image
Escape	Stop execution
Ctrl+Z	Undo
Ctrl+Y	Redo
Command Line Usage
bash
# Run a .rainbow file
rb.exe -lgbt script.rainbow

# Execute code from command line
rb.exe -c "comingout \"Hello, world!\";"

# Compile to executable
rb.exe -b input.rainbow output.exe

# Show tokens
rb.exe -tokens script.rainbow

# Show AST
rb.exe -ast script.rainbow
🌍 Web Server Features
LGBTScript includes a full-featured web server module:

rainbow
@ Server management
createServer(name, port)          @ Create a new server
startServer(name)                 @ Start the server
stopServer(name)                  @ Stop the server
addRoute(server, method, path, handler) @ Add HTTP route
getServerStatus(name)             @ Get server status
listServers()                     @ List all servers
Route Handlers
Route handlers receive query and body parameters:

query: Array of key-value pairs from URL parameters

body: Request body (for POST/PUT requests)

🏳️‍🌈 LGBTQ+ Resource Database
The getLGBTResources() function returns a JSON array of verified LGBTQ+ organizations with:

Name: Organization name

Type: network, psychological, advocacy, international, crisis, media

Description: Organization description

Country: Country location

City: City location

Address: Physical address (if available)

Phone: Contact phone number

Email: Contact email

Website: Organization website

Services: List of services offered

Working Hours: Operating hours

Verified: Boolean indicating verification status

Rating: Community rating

Languages: Supported languages

Example Response
json
{
  "total": 6,
  "resources": [
    {
      "id": "r1",
      "name": "Российская ЛГБТ-сеть",
      "type": "network",
      "description": "Крупнейшая российская ЛГБТ-организация...",
      "country": "Россия",
      "city": "Москва",
      "phone": "+7 (495) 123-45-67",
      "services": ["юридическая_помощь", "психологическая_поддержка"],
      "verified": true,
      "rating": 4.8
    }
  ],
  "filters": {
    "country": "Россия",
    "type": "network"
  }
}
🖼️ Image Generation
The lgbtImg() function generates LGBTQ+-themed images:

rainbow
@ Basic usage
lgbtImg("rainbow pride flag", 1024, 768);

@ With custom prompt
lgbtImg("diverse LGBTQ+ people celebrating pride parade, colorful, joy", 1024, 768);

@ The function returns information about the generated file:
@ - Filename
@ - Size
@ - Prompt used
@ - Seed
@ - Timestamp
📦 Compilation
LGBTScript can compile .rainbow files to standalone Windows executables:

bash
rb.exe -b script.rainbow app.exe
The compiled executable embeds the script and runs it automatically when launched.

🔒 Security
LGBTScript includes a sandbox with security features:

File path restrictions: Blocks access to system directories (/etc, /proc, /sys, etc.)

URL restrictions: Blocks access to localhost and internal addresses

File size limits: Prevents loading huge files (max 10MB)

HTTP timeouts: 30-second timeout for HTTP requests

Recursion protection: Max recursion depth of 1000

🧪 Testing
rainbow
@ Run the test suite
rainbow test() {
    gay passed = 0;
    gay failed = 0;
    
    @ Test arithmetic
    cis (2 + 2 == 4) {
        passed++;
    } nocis {
        failed++;
        comingout "❌ Math test failed!";
    }
    
    @ Test string concatenation
    cis ("Hello" + " World" == "Hello World") {
        passed++;
    } nocis {
        failed++;
        comingout "❌ String test failed!";
    }
    
    comingout "✅ Tests passed: " + passed;
    comingout "❌ Tests failed: " + failed;
}

test();
🤝 Contributing
We welcome contributions from everyone! Here's how you can help:

Report bugs: Open an issue on GitHub

Suggest features: Propose new LGBTQ+ functions or language features

Write documentation: Improve the docs or translate them

Submit code: Pull requests for bug fixes or new features

Development Setup
Fork the repository

Clone your fork

Open in PureBasic IDE

Make your changes

Test thoroughly

Submit a pull request

📚 Documentation
Full documentation is available at: lgbtscript.dev

🏳️‍🌈 Community
Discord: Join our community

Reddit: r/LGBTScript

Twitter: @LGBTScript

📄 License
This project is licensed under the Apache License, Version 2.0.

text
Copyright 2024 LGBTScript Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at:

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
💖 Support
If you or someone you know needs support, these resources are available:

Crisis Hotline: 8-800-XXX-XX-XX (24/7)

Online Chat: https://lgbt-support.org/chat

Resource Database: Use getLGBTResources() in LGBTScript

<div align="center"> <p>🏳️‍🌈 Love is Love. Code with Pride. 🏳️‍🌈</p> <p><sub>LGBTScript v9.0 "Pride Edition"</sub></p> </div> ```
🌟 Summary of Key Features
Category	Features
Language	Inclusive keywords, strong typing, OOP, control flow, error handling
Web	HTTP servers, routing, request handling
LGBTQ+	Resource database, crisis support, health info, community tools
Creative	Image generation, rainbow themes, pride visuals
IDE	Syntax highlighting, code intelligence, debugging, compilation
Security	Sandbox, file restrictions, URL blocking, timeouts
This documentation provides a comprehensive guide to LGBTScript and its IDE, covering all features, functions, examples, and the project's inclusive philosophy.

