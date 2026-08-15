# 🏳️‍🌈 LGBTScript & LGBTScript IDE v9.0

**The Inclusive Programming Language for the LGBTQ+ Community**

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Version](https://img.shields.io/badge/version-9.0-brightgreen)](https://github.com/your-username/lgbt-script/releases)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)

---

## 📖 Our Philosophy

LGBTScript is a high-level, interpreted programming language created **by and for the LGBTQ+ community**. It is designed to be a safe, empowering, and expressive tool for everyone—whether you are a seasoned developer or just starting your coding journey.

We believe that technology should be a space of inclusion, joy, and resistance. LGBTScript’s syntax and built-in functions are crafted to reflect our identities, celebrate our history, and provide practical support for our daily lives.

> **“Be yourself. Write code. Change the world.”**

---

## ✨ Key Features

- **🏳️‍🌈 Inclusive Syntax**: Keywords like `lesbian`, `gay`, `nonbinary`, `gender`, and `asexual` are integrated as first-class data types.
- **🛡️ Advanced Hate Speech Filter**: A real-time, multi-lingual filter (20+ languages) with morphological analysis to detect and neutralize hate speech.
- **🆕 18+ Support & Activism Functions**: From writing a coming-out letter (`comingOutLetter`) to finding a mentor (`findMentor`), LGBTScript is packed with tools for community support and personal growth.
- **🖼️ Built-in Image Generation**: Use the `lgbtImg()` function to generate custom AI art of pride flags, events, and more.
- **💬 Native GUI Support**: The `msg()` function creates native Windows dialog boxes, allowing you to build interactive applications.
- **📦 Standalone Compilation**: Use the IDE or the command line to compile your `.rainbow` scripts into executable `.exe` files that run on any Windows machine.

---

## 🚀 Quick Start

### Installation

1.  **Download the latest release** from the [Releases](https://github.com/your-username/lgbt-script/releases) page.
2.  Extract the contents to a folder of your choice.
3.  (Optional) Run `rb.exe` from the command line to test the interpreter.

### Your First Code

Create a file named `hello.rainbow` and paste the following:

```rainbow
@ My first LGBTScript program

lesbian message = "Hello, beautiful world! 🏳️‍🌈";
comingout message;
Run it using the command line:

bash
rb.exe -lgbt hello.rainbow
Or simply open it in the LGBTScript IDE and press the ▶ Run button.

📚 Language Guide
Data Types
Keyword	Type	Description	Example
lesbian	String	Text strings	"Love is love"
gay	Integer	Whole numbers	42, -7, 0
queer	Float	Floating-point numbers	3.1415, -0.001
nonbinary	Boolean	True or False	true, false
gender	Array	Ordered list of any type	["red", "orange", "yellow"]
asexual	Constant	Constant value (cannot change)	asexual pi = 3.14;
Core Functions
msg(title, text, [type], [icon])
Displays a native Windows MessageBox.

title: The dialog window title.

text: The message body.

type (optional): Button styles: "ok", "okcancel", "yesno", "yesnocancel", "retrycancel", "abortretryignore".

icon (optional): Icon types: "info", "warning", "error", "question".

Example:

rainbow
msg("Welcome", "Hello, LGBTQ+ community!", "ok", "info");
lgbtImg(prompt, width, height)
Generates an AI image based on your prompt and saves it as a PNG file.

prompt: Text description of the image.

width: Image width in pixels.

height: Image height in pixels.

Example:

rainbow
lgbtImg("A non-binary person under a rainbow flag at a pride parade", 1024, 768);
The #inclusive Directive
Import other LGBTScript files to organize your code.

Example:

rainbow
#inclusive "lib/support_functions.rainbow";
🆕 Support & Activism Library (18+ Functions)
These functions make LGBTScript unique. They are designed to provide emotional support, education, and tools for activism.

Function	Description
comingOutLetter(name, recipient, relationship, [style])	Generates a personalized coming-out letter.
transSupport(query)	Provides information on trans-related topics (hormones, documents, etc.).
findAllies(place, radius)	Lists LGBTQ+-friendly spaces near a given location.
lgbtHistory([year], [country])	Displays historical LGBTQ+ events.
prideAvatar(name, style)	Creates an ASCII pride avatar with a custom flag.
lgbtJournal(action, text, [mood])	Encrypted personal journal to store private thoughts.
getLGBTCEvents([country])	Returns a list of upcoming LGBTQ+ events.
createFlag(type, [size])	Renders a pride flag (e.g., "trans", "lesbian", "pride") in the console.
lgbtTherapy(exercise, [duration])	Offers therapeutic exercises for self-acceptance and anxiety.
lgbtQuote([theme])	Displays an inspiring LGBTQ+ quote.
lgbtBookClub([title])	Finds books with LGBTQ+ themes.
checkPrivilege(country)	Rates safety levels for LGBTQ+ people in a specific country.
createSafeSpaceCert(school, address)	Generates a certificate for an LGBTQ+ safe space.
activismPlan(goal, country)	Creates a step-by-step activism plan.
prideMandalas([name], [colors])	Generates a beautiful mandala of support.
lgbtTimeline()	A detailed timeline of the LGBTQ+ rights movement.
hateMap([region])	Displays a map of hate speech levels.
findMentor(interest, [experience])	Finds a mentor for a specific topic (e.g., coming out, career).
🛡️ Anti-Hate Filter
LGBTScript includes an invisible guardian: the Hate Filter. It runs in the background, scanning all output and input for hate speech in over 20 languages. It uses advanced morphological analysis to catch slurs even when they are disguised.

Action: When detected, the filter can either log the incident, warn the user, or automatically replace the hateful content with 🏳️‍🌈[заблокировано]🏳️‍🌈.

Logs: All incidents are saved to hate_speech_log.txt in the working directory.

Control: You can manage the filter with enableHateFilter() and disableHateFilter().

🖥️ LGBTScript IDE
The LGBTScript IDE is a native Windows application that makes writing and executing LGBTScript code easy and enjoyable.

Modern Dark Theme: A VS Code-inspired interface that is easy on the eyes.

Syntax Highlighting: Full support for LGBTScript's unique keywords and types.

Integrated Terminal: Output is displayed in a dedicated log panel.

One-Click Compilation: Compile your .rainbow scripts into standalone Windows .exe files.

Image Generation: A dedicated button to generate LGBTQ+ themed images via the Pollinations.ai API.

Error Navigation: Double-click on an error in the log to jump directly to the offending line of code.

🔧 Command Line Interface (CLI)
The interpreter (rb.exe) can be used directly from the command line.

Command	Description
rb.exe -lgbt <file.rainbow>	Execute a LGBTScript file.
rb.exe -c "<code>"	Execute a single line of code.
rb.exe -exe <input.rainbow> <output.exe>	Compile a script into a standalone executable.
rb.exe -tokens <file.rainbow>	Display the tokens from the lexer.
rb.exe -ast <file.rainbow>	Display the Abstract Syntax Tree.
rb.exe -example	Run the built-in example script.
Example:

bash
rb.exe -c "gay x = 10; comingout x * 2;"
# Output: 20
📦 Building an Executable
You can easily turn any LGBTScript into a standalone Windows executable.

Using the IDE:

Open your .rainbow file.

Click the "⚙ Компиляция" button.

Choose where to save your .exe file.

Using the CLI:

bash
rb.exe -exe my_script.rainbow my_app.exe
The resulting .exe file will run on any Windows machine without needing the interpreter installed.

📄 License
LGBTScript and LGBTScript IDE are open-source software licensed under the Apache License, Version 2.0.

text
Copyright (c) 2024-2026 LGBTScript Community

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
🤝 How to Contribute
We welcome contributions from everyone! Whether it's fixing a bug, adding a new feature, improving documentation, or simply using the language and reporting issues, your help is invaluable.

Fork the repository.

Create a branch for your feature (git checkout -b feature/amazing-feature).

Commit your changes (git commit -m 'Add some amazing feature').

Push to your branch (git push origin feature/amazing-feature).

Open a Pull Request.

Please read our Contributing Guidelines for more details.

🌈 Community & Support
GitHub Issues: For bug reports and feature requests.

Discord: Join our community server for support and discussion. (Link coming soon)

Email: lgbt-script@community.org

💖 Support the Project
If you like LGBTScript, please consider giving us a ⭐ on GitHub! It helps the project grow and reach more people.

Made with ❤️ and 🌈 by the LGBTScript Community.
