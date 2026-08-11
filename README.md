# 🌈 LGBTScript

> An inclusive, expressive programming language for everyone  
> *"Love is code that doesn't need compilation"*

[![Version](https://img.shields.io/badge/version-2.0.0-brightgreen.svg)](https://github.com/yourusername/lgbtscript)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)
[![Made with Love](https://img.shields.io/badge/Made%20with-❤️-ff69b4.svg)](https://github.com/yourusername/lgbtscript)
[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat&logo=go)](https://golang.org)

---

**LGBTScript** is an interpreted programming language built with inclusivity, friendliness, and expressiveness in mind. Keywords are inspired by LGBTQ+ community themes, making code not only functional but also symbolic. With built-in social support functions, educational tools, and community resources, LGBTScript empowers developers to create applications that support and celebrate diversity.

## ✨ Features

- 🔤 **5 primitive types**: `lesbian` (string), `gay` (integer), `trans` (float), `nonbinary` (boolean), `gender` (array)
- 📦 **Dynamic arrays** with flexible sizing
- 🧩 **Functions** with export support (`export rainbow`)
- 🔄 **Control flow**: `cis` (if), `nocis` (else if/else), `pride` (while)
- 📂 **Built-in functions** for file I/O, HTTP, JSON, cryptography, regex, and system operations
- 📚 **Library imports** via `#intersex`
- 🛡️ **Error handling** with `try`/`catch`
- 🌈 **Social support functions** - Crisis support, mental health, education
- 🏥 **Medical information** - HRT, LGBTQ+ friendly doctors
- 🎉 **Community tools** - Events, groups, volunteering
- 🎨 **Cultural resources** - Books, playlists, movies
- 🔒 **Sandbox security** - Safe execution environment
- 🧵 **Thread-safe** - Built-in concurrency safety

---

## 🚀 Quick Start

### Installation

Download the interpreter `rb.exe` or build from source:

```bash
git clone https://github.com/yourusername/lgbtscript.git
cd lgbtscript
go build -o rb.exe main.go
Your First Program
Create a file hello.rainbow:

go
@ First LGBTScript program

lesbian name = "World";
comingout "Hello, " + name + "! 🌈";
Run it:

bash
rb.exe hello.rainbow
Or execute code directly:

bash
rb.exe -c 'comingout "🌈 Hello from LGBTScript!";'
🏷️ Data Types
LGBTScript uses inclusive keywords for type declarations:

Type	Keyword	Description	Default Value
String	lesbian	Text data	""
Integer	gay	Whole numbers	0
Float	trans	Decimal numbers	0.0
Boolean	nonbinary	true or false	false
Array	gender	Collection of values	[]
Literals
go
"Hello, World!"   @ String
42                @ Integer
3.14              @ Float
true              @ Boolean
[1, 2, 3]         @ Array
📦 Variables
Variables are declared with a type. Default values depend on the type:

go
@ Declaration with initialization
lesbian name = "Alice";
gay age = 30;
trans pi = 3.14159;
nonbinary isReady = true;
gender hobbies = ["reading", "coding", "gaming"];

@ Declaration without initialization (gets default value)
lesbian message;
comingout message;  @ outputs empty string

@ Assignment (type must match)
name = "Bob";
age = age + 1;
hobbies[0] = "swimming";
➕ Operators
Arithmetic
text
+  -  *  /  %
Comparison
text
==  !=  <  >  <=  >=
Logical
text
&&  ||
String Concatenation
The + operator works with strings too:

go
lesbian greeting = "Hello, " + "World!";
lesbian full = "Value: " + 42;  @ "Value: 42"
📊 Arrays
Arrays can hold any values, including mixed types.

go
gender numbers = [1, 2, 3, 4, 5];
gender fruits = ["apple", "banana", "cherry"];
gender mixed = ["text", 42, true, 3.14];

@ Index access
comingout numbers[0];  @ 1
comingout mixed[1];    @ 42

@ Index assignment
numbers[2] = 99;

@ Built-in array functions
append(numbers, 6);        @ adds an element
gay len = length(numbers); @ array length
remove(numbers, 0);        @ removes element at index
🔄 Control Flow
Conditional Statement cis / nocis
go
cis (age >= 18) {
    comingout "You are an adult";
} nocis (age >= 13) {
    comingout "You are a teenager";
} nocis {
    comingout "You are a child";
}
Loop pride
go
gay i = 0;
pride (i < 5) {
    comingout "Count: " + i;
    i = i + 1;
}
🧩 Functions
Functions are declared with the rainbow keyword. Export with export.

go
@ Simple function
rainbow greet(name) {
    return "Hello, " + name + "!";
}

@ Exported function
export rainbow calculate(a, b) {
    return a + b;
}

@ Function call
lesbian msg = greet("Alice");
comingout msg;
🛡️ Error Handling
The try / catch construct catches errors. The error variable is available in the catch block.

go
try {
    lesbian result = httpGet("https://invalid.url");
    comingout result;
} catch {
    comingout "Error: " + error;
}
📚 Libraries
Import external libraries using the #intersex directive:

go
@ Import libraries
#intersex "math.rainbow"
#intersex "io.rainbow"
Library search paths:

Current directory

Directory of the executable

libs and libraries folders

🧰 Built-in Functions
🌈 Social & Support Functions
Function	Description	Example
findSafeSpace(place, city, radius)	Finds LGBTQ+ friendly places nearby	findSafeSpace("cafe", "Moscow", 5)
getCrisisSupport(region, type)	Gets crisis support contacts	getCrisisSupport("Russia", "hotline")
getLGBTQLaws(country, category)	Gets information about LGBTQ+ laws	getLGBTQLaws("USA", "rights")
getDailyAffirmation(theme)	Gets a daily affirmation	getDailyAffirmation("self-love")
moodCheck(moods, suggestResources)	Checks emotional state	moodCheck(["anxiety"], true)
guidedBreathing(minutes, theme)	Guided breathing exercises	guidedBreathing(5, "calm")
defineTerm(term, language)	Defines LGBTQ+ terms	defineTerm("nonbinary", "ru")
lgbtHistoryQuiz(difficulty)	LGBTQ+ history quiz	lgbtHistoryQuiz("medium")
getDailyFact(category, region)	Daily LGBTQ+ fact	getDailyFact("culture", "global")
getHRTInfo(country, type)	Hormone replacement therapy info	getHRTInfo("USA", "MTF")
findLGBTDoctor(specialty, city)	Find LGBTQ+ friendly doctors	findLGBTDoctor("therapist", "Berlin")
getDocumentChangeGuide(country, document)	Guide for changing documents	getDocumentChangeGuide("UK", "passport")
getLGBTQEvents(days, type)	Upcoming LGBTQ+ events	getLGBTQEvents(30, "online")
createLGBTQGroup(name, meetingType)	Creates an LGBTQ+ group	createLGBTQGroup("Book Club", "online")
findVolunteerOpportunity(organization, skills)	Finds volunteer opportunities	findVolunteerOpportunity("Rainbow Org", ["design"])
getLGBTQBook(genre)	LGBTQ+ book recommendations	getLGBTQBook("fantasy")
getLGBTQPlaylist(mood)	LGBTQ+ playlist recommendations	getLGBTQPlaylist("empowerment")
getLGBTQMovies(genre)	LGBTQ+ movie recommendations	getLGBTQMovies("romance")
📁 File System
Function	Description
readFile(filename)	Reads file, returns string
writeFile(filename, content)	Writes data to file
fileExists(filename)	Checks if file exists
getDirFiles(path)	Returns list of files in directory
🔤 String Operations
Function	Description
split(text, delimiter)	Splits string into array
replace(text, old, new)	Replaces substrings
trim(text)	Removes leading/trailing whitespace
length(value)	Length of string or array
toUpper(text) / toLower(text)	Case conversion
🔍 Regular Expressions
Function	Description
regexFind(pattern, text)	Finds all matches
regexReplace(pattern, text, replacement)	Replaces using pattern
🌐 HTTP
Function	Description
httpGet(url)	GET request
httpPost(url, data)	POST request (JSON)
📊 JSON
Function	Description
jsonParse(jsonString)	Parses JSON into object/array
🔐 Cryptography
Function	Description
md5(text)	MD5 hash
sha256(text)	SHA256 hash
💻 System
Function	Description
getTime()	Current date and time
getYear()	Current year
getMonth()	Current month (1–12)
getOS()	Operating system name
getHostname()	Hostname
getArgs()	Command-line arguments
hasFlag(flag)	Checks if flag is present
📐 Mathematics
Function	Description
random(min, max)	Random integer
max(...) / min(...)	Maximum/minimum
sqrt(number)	Square root
pow(base, exp)	Power
🛠️ Utilities
Function	Description
sleep(ms)	Sleep in milliseconds
append(array, element)	Adds element to array
remove(array, index)	Removes element at index
sendEmail(to, subject, body)	Email simulation
help(country)	LGBTQ+ organization support
orientation()	Interactive orientation test
💻 Command Line Interface
bash
rb.exe [options] [file]

Options:
  -c "code"         Execute code from command line
  -lgbt file        Execute file (alternative to positional argument)
  -tokens           Show tokens after lexical analysis
  -ast              Show AST after parsing
  -debug            Enable debug mode
  -example          Show extended example with social functions

Examples:
  rb.exe script.rainbow
  rb.exe -c 'comingout "Hello!";'
  rb.exe -lgbt script.rainbow -tokens
  rb.exe --example
📝 Examples
Basic Example
go
@ Simple program
lesbian name = "World";
comingout "Hello, " + name + "! 🌈";

gay x = 10;
gay y = 20;
comingout "Sum: " + (x + y);
Social Support Example
go
@ LGBTQ+ Support Program

rainbow main() {
    comingout "🌈 Welcome to LGBTScript Support!";
    
    @ Get crisis support
    lesbian support = getCrisisSupport("Russia", "hotline");
    comingout support;
    
    @ Daily affirmation
    lesbian affirmation = getDailyAffirmation("self-love");
    comingout affirmation;
    
    @ Check mood
    gender moods = ["anxiety", "loneliness"];
    lesbian moodResult = moodCheck(moods, true);
    comingout moodResult;
    
    @ Find safe spaces
    lesbian spaces = findSafeSpace("cafe", "Moscow", 5);
    comingout spaces;
    
    @ Get book recommendation
    lesbian book = getLGBTQBook("fantasy");
    comingout book;
    
    return 0;
}

main();
File and HTTP Operations
go
#intersex "io.rainbow"

lesbian url = "https://api.github.com";
try {
    lesbian response = httpGet(url);
    writeFile("github.json", response);
    comingout "Data saved";
} catch {
    comingout "Error: " + error;
}
Functions and Arrays
go
export rainbow processArray(arr) {
    gay sum = 0;
    gay i = 0;
    pride (i < length(arr)) {
        sum = sum + arr[i];
        i = i + 1;
    }
    return sum;
}

gay numbers = [1, 2, 3, 4, 5];
gay total = processArray(numbers);
comingout "Sum: " + total;
Extended Example
go
@ Extended LGBTScript example

@ Import libraries
#intersex "math.rainbow"
#intersex "io.rainbow"

@ Exported function
export rainbow processData(data) {
    comingout "Processing: " + data;
    return "Processed: " + data;
}

@ Main program
rainbow main() {
    @ File operations
    lesbian config = readFile("config.json");
    comingout "Config loaded: " + config;
    
    @ JSON parsing
    lesbian settings = jsonParse(config);
    comingout "API URL: " + settings["api_url"];
    
    @ HTTP request
    try {
        lesbian response = httpGet(settings["api_url"]);
        comingout "Response received";
        writeFile("response.txt", response);
    } catch {
        comingout "Error: " + error;
    }
    
    @ Array operations
    gender numbers = [1, 2, 3, 4, 5];
    append(numbers, 6);
    comingout "Numbers: " + numbers;
    comingout "First element: " + numbers[0];
    
    @ String operations
    lesbian text = "  Hello World  ";
    lesbian clean = trim(text);
    lesbian upper = toUpper(clean);
    comingout "Upper: " + upper;
    
    @ Regular expressions
    gay matches = regexFind("\\d+", "Phone: 123-456-7890");
    comingout "Found numbers: " + matches;
    
    @ Cryptography
    lesbian hash = md5("password");
    comingout "MD5: " + hash;
    
    @ System information
    lesbian os = getOS();
    lesbian host = getHostname();
    comingout "OS: " + os + ", Host: " + host;
    
    @ Command-line arguments
    gay args = getArgs();
    comingout "Arguments: " + args;
    
    @ Social support
    lesbian support = getCrisisSupport("global", "hotline");
    comingout support;
    
    lesbian affirmation = getDailyAffirmation("strength");
    comingout affirmation;
    
    lesbian book = getLGBTQBook("fantasy");
    comingout "Recommended book: " + book;
    
    nonbinary verbose = hasFlag("--verbose");
    cis (verbose) {
        comingout "Verbose mode enabled";
    }
    
    return 0;
}

@ Run
main();
Help Command — LGBTQ+ Organization Support
go
help "russia";   @ Shows organizations in Russia
help "usa";      @ Shows organizations in the USA
help "all";      @ Shows international organizations
Orientation Command — Orientation Test
go
orientation;     @ Runs an interactive orientation test
🔒 Security Features
LGBTScript includes built-in security features:

Sandbox Environment: Restricted file system access

Blocked Paths: /etc, /proc, /sys, /root, /home

File Size Limits: Maximum 10MB file size

HTTP Timeout: 30 second timeout for requests

Recursion Limit: Maximum 1000 recursive calls

Domain Blocking: Localhost and 127.0.0.1 blocked for HTTP

Thread-Safe: Mutex protection for concurrent execution

🤝 Contributing
We welcome contributions! Please see our Contributing Guide for details.

Fork the repository

Create your feature branch (git checkout -b feature/AmazingFeature)

Commit your changes (git commit -m 'Add some AmazingFeature')

Push to the branch (git push origin feature/AmazingFeature)

Open a Pull Request

📄 License
Copyright 2024 LGBTScript Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

🌈 Community
🐛 Report an issue

💡 Suggest a feature

📚 Documentation

💬 Discord Community

<div align="center"> 🌈 LGBTScript — a language where everyone can find themselves
Made with love for everyone

⬆ Back to top

</div> ```
