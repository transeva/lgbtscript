# 🌈 LGBTScript

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat&logo=go)](https://golang.org/)

**LGBTScript** is an interpreted programming language designed with inclusivity at its core. It features a unique, pride-inspired syntax, built-in functions for LGBTQ+ support, and capabilities for web server development. It aims to be a welcoming and safe space for everyone to learn and create.

## ✨ Features

- 🏳️‍🌈 **Inclusive Syntax:** Uses keywords like `lesbian`, `gay`, `trans`, `nonbinary`, and `gender` for variable types.
- 🛠️ **Built-in Social & Support Functions:** Access resources for crisis support, find safe spaces, learn about LGBTQ+ history, and more directly from your code.
- 🌐 **Web Server Capabilities:** Create, configure, and run your own web servers using built-in functions.
- 🧩 **Object-Oriented Programming:** Full support for classes (`queer`), inheritance (`extends`), methods, and the `this`/`super` keywords.
- 🔒 **Secure & Sandboxed:** Includes a sandbox environment to prevent unsafe file system and network access.
- ❤️ **Community Focused:** Built as a safe and educational space for the LGBTQ+ community and allies.

## 🚀 Getting Started

### Installation

You need Go (1.20+) installed to build LGBTScript.

```bash
git clone https://github.com/your-username/lgbt-script.git
cd lgbt-script
go build -o lgbt
This will create the lgbt executable in your current directory.

Hello, World!
Create a file named hello.rainbow:

ruby
@ Hello World program
rainbow main() {
    comingout "Hello, World! 🌈";
}

main();
Run it:

bash
./lgbt hello.rainbow
📖 Language Guide
Variable Declarations
Keyword	Data Type	Description
lesbian	String	Text strings
gay	Int	Integer numbers
trans	Float	Floating-point numbers
nonbinary	Bool	Boolean values
gender	Array	Lists/Arrays
Example:

ruby
lesbian name = "Alex";
gay age = 30;
trans height = 5.9;
nonbinary is_student = true;
gender colors = ["red", "green", "blue"];
Control Structures
If-Else: cis (condition) { ... } nocis (condition) { ... }

While Loop: pride (condition) { ... }

For Loop: sex (init; condition; update) { ... }

Example of a For Loop:

ruby
sex (gay i = 0; i < 5; i++) {
    comingout "Number: " + i;
}
Functions
Define functions using the rainbow keyword.

ruby
rainbow add(gay a, gay b) {
    return a + b;
}

gay sum = add(5, 3);
comingout sum;  # Outputs 8
Object-Oriented Programming
Define classes with queer. Use extends for inheritance.

ruby
queer Person {
    lesbian name;
    gay age;

    rainbow init(name, age) {
        this.name = name;
        this.age = age;
    }

    rainbow greet() {
        comingout "Hello, my name is " + this.name;
    }
}

queer p = new Person("Taylor", 25);
p.greet();
Error Handling
Use try and catch for errors.

ruby
try {
    lesbian content = readFile("missing.txt");
} catch {
    comingout "File not found!";
}
📚 Built-in Functions
LGBTScript includes a variety of built-in functions:

File I/O: readFile(), writeFile(), fileExists()

String Manipulation: split(), replace(), trim(), length(), toUpper(), toLower()

HTTP Client: httpGet(), httpPost()

JSON: jsonParse()

Math: random(), max(), min(), sqrt(), pow()

System: getTime(), getOS(), getArgs(), hasFlag()

Server Management: createServer(), startServer(), stopServer(), addRoute()

LGBTQ+ Support: findSafeSpace(), getCrisisSupport(), getDailyAffirmation(), lgbtHistoryQuiz(), getHRTInfo(), findLGBTDoctor(), and many more.

Example:

ruby
lesbian data = readFile("data.txt");
comingout "File content: " + data;
📄 License
This project is licensed under the Apache License, Version 2.0 - see the LICENSE file for details.

text
Copyright [year] [your name]

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
🤝 Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

🌈 Support
If you or someone you know needs support, please reach out to local LGBTQ+ organizations or crisis hotlines. You are not alone.

LGBTScript - Programming with Pride.
