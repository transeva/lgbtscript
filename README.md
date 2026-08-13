markdown
# 🏳️‍🌈 LGBTScript & LGBTScript IDE

**An Inclusive, Expressive, and Safe Programming Language for the LGBTQ+ Community**

![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)
![Version](https://img.shields.io/badge/version-7.0-ff69b4.svg)
![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)
![Made with Love](https://img.shields.io/badge/Made%20with-Love-ff69b4.svg)

LGBTScript is a programming language designed to be a validating, empowering, and safe environment for the LGBTQ+ community. It is not just a tool; it's a statement that everyone deserves to code with pride, respect, and safety.

[Report Bug](https://github.com/your-repo/issues) · [Request Feature](https://github.com/your-repo/issues) · [View Documentation](https://your-docs-url.com)

---

## 🌈 Philosophy

LGBTScript was built with a clear purpose:

- **Validate Identity:** Syntax and keywords reflect LGBTQ+ identities and themes.
- **Promote Safety:** A built-in Hate Filter neutralizes hate speech in real-time across 20+ languages.
- **Empower Communication:** Create safe, filtered WebSocket chat servers for community building.
- **Be Accessible:** Human-readable code and an inclusive IDE make programming approachable for everyone.
- **Remain Open & Free:** Licensed under **Apache 2.0**, ensuring it is available for all.

## ✨ Core Features

- **Inclusive Data Types:** `lesbian`, `gay`, `queer`, `nonbinary`, `gender`
- **Safe by Design:** Real-time Hate Filter with logging and masking.
- **WebSocket Chat:** Create secure, filtered chat servers with a web UI.
- **Native GUI:** Build Windows desktop applications with ease.
- **HTTP Server:** Create and manage RESTful web servers.
- **Object-Oriented Programming:** Full support for classes, inheritance, and more.
- **Social Functions:** Access resources, support, and affirmations.
- **Compilable:** Create standalone `.exe` files for your scripts.

---

## 🚀 Getting Started

### Prerequisites

- **Go 1.19+**: Required to build from source.
- **Windows OS**: Full support for GUI and most native functions. Other OSes are supported for core scripting.

### Installation

#### Option 1: Download a Release

Download the latest release from the [Releases Page](https://github.com/your-repo/releases).

#### Option 2: Build from Source

```bash
git clone https://github.com/your-repo/lgbtscript.git
cd lgbtscript
go build -o lgbt.exe main.go
Your First Program
Create a file called hello.rainbow:

go
rainbow main() {
    comingout "Hello, LGBTScript! 🏳️‍🌈";
}
main();
Run it:

bash
./lgbt.exe hello.rainbow
📚 Language Basics
Data Types & Declarations
Keyword	Type	Description	Example
lesbian	String	Text values	lesbian name = "Alex";
gay	Integer	Whole numbers	gay age = 25;
queer	Float	Decimal numbers	queer score = 98.6;
nonbinary	Boolean	True/False	nonbinary isMember = true;
gender	Array	List of values	gender list = [1, 2, 3];
asexual	Constant	Immutable value	asexual PI = 3.14159;
Control Flow
go
// Conditional (cis)
cis age >= 18 {
    comingout "Adult";
} nocis age >= 13 {
    comingout "Teen";
} cis {
    comingout "Child";
}

// While Loop (pride)
gay i = 0;
pride i < 5 {
    comingout i;
    i = i + 1;
}

// For Loop (sex)
sex (gay i = 0; i < 5; i = i + 1) {
    comingout i;
}
Functions
go
rainbow greet(lesbian name) {
    comingout "Hello, " + name + "!";
}

export rainbow add(gay a, gay b) {
    return a + b;
}
Queer Classes (OOP)
go
queer Person {
    lesbian name;
    gay age;

    rainbow constructor(lesbian n, gay a) {
        this.name = n;
        this.age = a;
    }
}

queer Employee extends Person {
    lesbian role;
    // ...
}
🛡️ Hate Filter
The Hate Filter is always active in the interpreter, chat servers, and GUI.

go
lesbian text = "This is a slur: faggot";
lesbian result = checkHate(text);
comingout result; // JSON with 'has_hate': true
Functions:

checkHate(text): Checks for hate speech.

filterHate(text): Returns filtered text.

enableHateFilter() / disableHateFilter(): Toggle the filter.

addHateSlur(lang, slur), removeHateSlur(lang, slur): Customize the filter.

getHateLog(), clearHateLog(), getHateStats(): Manage logs.

💬 WebSocket Chat
Create safe, real-time chat servers with built-in filtering.

go
createLGBTChat("mainChat", 8080, 100);
startLGBTChat("mainChat", 8080);
sendLGBTChatMessage("mainChat", "Bot", "Hello, world!");
Functions:

createLGBTChat(name, port, maxMessages)

startLGBTChat(name, port)

stopLGBTChat(name)

sendLGBTChatMessage(chatName, username, msg)

getLGBTChatMessages(chatName, limit)

getLGBTChatStats(chatName)

listLGBTChats()

🖥️ GUI Functions (Windows)
go
rainbowWin("main", "My App", 400, 300);
rainbowButton("main", "btn", "Click Me", 50, 50, 100, 30);

rainbow handleClick() {
    msg("Info", "Button clicked!", "ok", "info");
}

rainbowOnClick("main", "btn", "handleClick");
🌐 HTTP Server
go
createServer("myAPI", 8080);
addRoute("myAPI", "GET", "/hello", rainbow () {
    return "Hello, World!";
});
startServer("myAPI");
📦 Compiling to .exe
bash
# Standard executable
./lgbt.exe -b script.rainbow app.exe

# GUI executable (no console window)
./lgbt.exe -exe script.rainbow app_gui.exe
🖋️ LGBTScript IDE
The LGBTScript IDE is an integrated development environment designed for LGBTScript.

Features:

🌈 Themed UI: A welcoming, pride-themed interface.

📝 Code Editor: Syntax highlighting, auto-indentation, error detection.

▶️ Integrated Runner: Execute scripts and see output in real-time.

🛡️ Hate Filter Integration: Safe output pane.

📊 Project Management: Create, save, and manage .rainbow scripts.

📚 Built-in Documentation: Quick access to language docs and examples.

🔧 Compiler Support: Compile scripts to .exe with one click.

🌐 Built-in Chat Client: Test WebSocket chat servers.

📘 Examples
Hello World
go
rainbow main() {
    comingout "Hello, LGBTScript! 🏳️‍🌈";
}
main();
Chat Server
go
rainbow main() {
    createLGBTChat("community", 8080, 100);
    startLGBTChat("community", 8080);
    sendLGBTChatMessage("community", "System", "Chat is live!");
    comingout "Server running at http://localhost:8080";
}
main();
Social Functions
go
rainbow main() {
    lesbian resources = getLGBTResources("USA", "crisis", "New York");
    comingout resources;
}
main();
OOP Example
go
queer Animal {
    lesbian name;
    rainbow constructor(lesbian n) { this.name = n; }
    rainbow speak() { comingout "Animal sound"; }
}

queer Dog extends Animal {
    rainbow speak() { comingout "Woof!"; }
}

lesbian d = new Dog("Buddy");
d.speak();
🤝 Contributing
We welcome contributions of all kinds! Whether it's reporting a bug, suggesting a feature, or submitting a pull request, your help is appreciated.

Fork the repository.

Create your feature branch (git checkout -b feature/AmazingFeature).

Commit your changes (git commit -m 'Add some AmazingFeature').

Push to the branch (git push origin feature/AmazingFeature).

Open a Pull Request.

📄 License
LGBTScript and the LGBTScript IDE are released under the Apache License, Version 2.0. See the LICENSE file for details.

You are free to:

Use the software for any purpose, commercial or non-commercial.

Modify the software to suit your needs.

Distribute the software or your modifications.

You must:

Retain the copyright and license notices.

Include a copy of the license.

Clearly state any changes you have made.

A full copy of the license is available at https://www.apache.org/licenses/LICENSE-2.0.

🙌 Acknowledgments
The LGBTQ+ community for the inspiration and support.

The open-source community for the tools that made this possible.

Everyone who believes that technology should be inclusive and safe for all.

Built with Pride, for the Community.

https://img.shields.io/github/stars/your-repo/lgbtscript.svg?style=social

text
This response is AI-generated, for reference only.
