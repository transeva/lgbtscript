# 🌈 LGBTScript

> An inclusive, expressive programming language for everyone  
> *"Love is code that doesn't need compilation"*

[![Version](https://img.shields.io/badge/version-1.0.0-brightgreen.svg)](https://github.com/yourusername/lgbtscript)
[![License](https://img.shields.io/badge/license-APACHE-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)
[![Made with Love](https://img.shields.io/badge/Made%20with-❤️-ff69b4.svg)](https://github.com/yourusername/lgbtscript)



---

**LGBTScript** is an interpreted programming language built with inclusivity, friendliness, and expressiveness in mind. Keywords are inspired by LGBTQ+ community themes, making code not only functional but also symbolic.

### ✨ Features

- 🔤 **4 primitive types**: `lesbian` (string), `gay` (integer), `trans` (float), `nonbinary` (boolean)
- 📦 **Dynamic arrays** with flexible sizing
- 🧩 **Functions** with export support (`export rainbow`)
- 🔄 **Control flow**: `gender` (if), `queer` (else), `pride` (while)
- 📂 **Built-in functions** for file I/O, HTTP, JSON, cryptography, regex, and system operations
- 📚 **Library imports** via `#intersex`
- 🛡️ **Error handling** with `try`/`catch`

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

rainbow
@ First LGBTScript program

lesbian name = "World";
comingout "Hello, " + name + "!";
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
Literals
rainbow
"Hello, World!"   @ String
42                @ Integer
3.14              @ Float
true              @ Boolean
[1, 2, 3]         @ Array
📦 Variables
Variables are declared with a type. Default values depend on the type:

rainbow
@ Declaration with initialization
lesbian name = "Alice";
gay age = 30;
trans pi = 3.14159;
nonbinary isReady = true;

@ Declaration without initialization (gets default value)
lesbian message;
comingout message;  @ outputs empty string

@ Assignment (type must match)
name = "Bob";
age = age + 1;
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

rainbow
lesbian greeting = "Hello, " + "World!";
📊 Arrays
Arrays can hold any values, including mixed types.

rainbow
gay numbers = [1, 2, 3, 4, 5];
lesbian fruits = ["apple", "banana", "cherry"];

@ Index access
comingout numbers[0];  @ 1

@ Index assignment
numbers[2] = 99;

@ Built-in array functions
append(numbers, 6);        @ adds an element
gay len = length(numbers); @ array length
remove(numbers, 0);        @ removes element at index
🔄 Control Flow
Conditional Statement gender
rainbow
gender (age >= 18) {
    comingout "You are an adult";
} queer {
    comingout "You are a minor";
}
Loop pride
rainbow
gay i = 0;
pride (i < 5) {
    comingout "Count: " + i;
    i = i + 1;
}
🧩 Functions
Functions are declared with the rainbow keyword. Export with export.

rainbow
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
🧰 Built-in Functions
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
Usage Example
rainbow
@ File and HTTP operations
lesbian data = readFile("config.json");
lesbian config = jsonParse(data);
comingout "URL: " + config["api_url"];

lesbian response = httpGet(config["api_url"]);
writeFile("response.txt", response);

@ Cryptography
lesbian hash = md5(response);
comingout "MD5: " + hash;

@ System information
lesbian os = getOS();
lesbian host = getHostname();
comingout "OS: " + os + ", Host: " + host;
🛡️ Error Handling
The try / catch construct catches errors. The error variable is available in the catch block.

rainbow
try {
    lesbian result = httpGet("https://invalid.url");
    comingout result;
} catch {
    comingout "Error: " + error;
}
📚 Libraries
Import external libraries using the #intersex directive:

rainbow
@ Import libraries
#intersex "math.rainbow"
#intersex "io.rainbow"
Library search paths:

Current directory

Directory of the executable

libs and libraries folders

💻 Command Line Interface
bash
rb10.exe [options] [file]

Options:
  -c "code"         Execute code from command line
  -lgbt file        Execute file (alternative to positional argument)
  -tokens           Show tokens after lexical analysis
  -ast              Show AST after parsing
  -debug            Enable debug mode
  -example          Show extended example

Examples:
  rb10.exe script.rainbow
  rb10.exe -c 'comingout "Hello!";'
  rb10.exe -lgbt script.rainbow -tokens
📝 Examples
Basic Example
rainbow
@ Simple program
lesbian name = "World";
comingout "Hello, " + name + "!";

gay x = 10;
gay y = 20;
comingout "Sum: " + (x + y);
File and HTTP Operations
rainbow
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
rainbow
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
rainbow
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
    gay numbers = [1, 2, 3, 4, 5];
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
    
    nonbinary verbose = hasFlag("--verbose");
    if (verbose) {
        comingout "Verbose mode enabled";
    }
    
    return 0;
}

@ Run
main();
help Command — LGBTQ+ Organization Support
rainbow
help "russia";   @ Shows organizations in Russia
help "usa";      @ Shows organizations in the USA
help "all";      @ Shows international organizations
orientation Command — Orientation Test
rainbow
orientation;     @ Runs an interactive orientation test
🌈 Community
🐛 Report an issue

💡 Suggest a feature

📚 Documentation

📄 License
This project is licensed under the MIT License — see the LICENSE file for details.

<div align="center">
🌈 LGBTScript — a language where everyone can find themselves

Made with love for everyone

⬆ Back to top

</div> ```
