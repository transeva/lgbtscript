LGBTScript Documentation
Overview
LGBTScript is a modern, expressive programming language designed with inclusivity and simplicity in mind. It combines familiar programming concepts with unique, inclusive syntax inspired by LGBTQ+ terminology. The language is interpreted and supports a wide range of features including variable types, functions, control flow, arrays, error handling, and built-in libraries.

Language Features
Key Concepts
Type System: Strongly typed with four primary data types

Functions: First-class functions with parameter support

Error Handling: Try-catch blocks for graceful error management

Modularity: Import system for code reuse

Arrays: Dynamic arrays with built-in operations

Built-in Libraries: Rich standard library for common tasks

Data Types
Basic Types
Type	Keyword	Description	Default Value
String	lesbian	Text data	"" (empty string)
Integer	gay	Whole numbers	0
Float	trans	Decimal numbers	0.0
Boolean	nonbinary	True/False values	false
Type Examples
rainbow
lesbian name = "Alex";
gay age = 25;
trans height = 1.75;
nonbinary isStudent = true;
Syntax and Structure
Comments
Comments start with the @ symbol and continue to the end of the line:

rainbow
@ This is a single-line comment
Variables
Variables must be declared with their type before use:

rainbow
@ Declaration with initialization
lesbian message = "Hello World";

@ Declaration without initialization (uses default value)
gay count;
Assignments
rainbow
@ Assignment
name = "New Name";
age = age + 1;
Operators
Arithmetic Operators
Operator	Operation	Example
+	Addition/Concatenation	5 + 3, "Hello " + "World"
-	Subtraction	10 - 4
*	Multiplication	6 * 7
/	Division	15 / 3
%	Modulo	10 % 3
Comparison Operators
Operator	Description	Example
==	Equal to	x == 5
!=	Not equal to	x != 5
<	Less than	x < 10
>	Greater than	x > 10
<=	Less than or equal	x <= 10
>=	Greater than or equal	x >= 10
Logical Operators
Operator	Description	Example
&&	Logical AND	x > 0 && x < 10
||	Logical OR	x < 0 || x > 10
Control Flow
Conditional Statements
The gender keyword is used for if-else statements:

rainbow
gender (age >= 18) {
    comingout "You are an adult";
} queer {
    comingout "You are a minor";
}
While Loops
The pride keyword is used for while loops:

rainbow
gay counter = 0;
pride (counter < 5) {
    comingout "Counter: " + counter;
    counter = counter + 1;
}
Functions
Function Declaration
Use the rainbow keyword to declare functions:

rainbow
rainbow greet(name) {
    return "Hello, " + name + "!";
}

rainbow add(a, b) {
    return a + b;
}
Exporting Functions
Use the export keyword to make functions available to other files:

rainbow
export rainbow processData(data) {
    @ This function can be used by other files
    return "Processed: " + data;
}
Function Calls
rainbow
lesbian greeting = greet("Alice");
gay result = add(5, 3);
Error Handling
Try-Catch Blocks
rainbow
try {
    lesbian data = readFile("config.json");
    comingout "File loaded successfully";
} catch {
    comingout "Error loading file: " + error;
}
Arrays
Creating Arrays
rainbow
gay numbers = [1, 2, 3, 4, 5];
lesbian names = ["Alice", "Bob", "Charlie"];
Accessing Elements
rainbow
gay first = numbers[0];  @ Access first element
numbers[2] = 10;         @ Modify element
Array Operations
Built-in functions for array manipulation:

rainbow
append(names, "David");           @ Add element
gay size = length(numbers);       @ Get array length
remove(numbers, 2);               @ Remove element at index
Input/Output
Print Statements
Use the comingout keyword to print:

rainbow
comingout "Hello World";
comingout 42;
comingout "Value: " + x;
Import System
Including Libraries
Use the #intersex directive to import libraries:

rainbow
#intersex "math.rainbow"
#intersex "io.rainbow"
Library Search Paths
The interpreter searches for libraries in:

Current directory

Directory of the source file

libs/ directory

libraries/ directory

Files with .rb or .rainbow extensions

Built-in Functions
File Operations
Function	Description	Parameters	Return
readFile(filename)	Read file contents	string filename	string content
writeFile(filename, content)	Write to file	string filename, string content	nil
fileExists(filename)	Check if file exists	string filename	boolean
getDirFiles(directory)	List files in directory	string directory path	array of strings
String Operations
Function	Description	Parameters	Return
split(text, delimiter)	Split string	string, string	array
replace(text, old, new)	Replace substrings	string, string, string	string
trim(text)	Remove whitespace	string	string
length(value)	Get length	string or array	int
toUpper(text)	Convert to uppercase	string	string
toLower(text)	Convert to lowercase	string	string
HTTP Operations
Function	Description	Parameters	Return
httpGet(url)	GET request	string URL	string response
httpPost(url, data)	POST request	string URL, string data	string response
JSON Operations
Function	Description	Parameters	Return
jsonParse(jsonString)	Parse JSON	string JSON	object or array
Cryptography
Function	Description	Parameters	Return
md5(text)	MD5 hash	string	string hash
sha256(text)	SHA256 hash	string	string hash
Regular Expressions
Function	Description	Parameters	Return
regexFind(pattern, text)	Find matches	string pattern, string text	array of matches
regexReplace(pattern, text, replacement)	Replace matches	string pattern, string text, string replacement	string
System Information
Function	Description	Parameters	Return
getTime()	Current time	None	string timestamp
getYear()	Current year	None	int
getMonth()	Current month	None	int
getOS()	Operating system	None	string
getHostname()	Computer name	None	string
getArgs()	Command-line arguments	None	array of strings
hasFlag(flag)	Check for flag	string flag	boolean
Math Functions
Function	Description	Parameters	Return
random(min, max)	Random integer	int min, int max	int
max(numbers...)	Maximum value	numbers	number
min(numbers...)	Minimum value	numbers	number
sqrt(number)	Square root	number	float
pow(base, exponent)	Power	number base, int exponent	number
Other Utilities
Function	Description	Parameters	Return
sleep(milliseconds)	Sleep	int ms	nil
sendEmail(to, subject, body)	Send email	string to, string subject, string body	nil
Special Commands
Help Command
The help keyword displays LGBTQ+ support organizations:

rainbow
help "russia";   @ Show organizations in Russia
help "usa";      @ Show organizations in USA
help "all";      @ Show international organizations
Orientation Test
The orientation keyword runs a psychological orientation test:

rainbow
orientation;  @ Runs interactive test
Command-Line Usage
Running Programs
bash
rainbow program.rainbow
Command-Line Options
bash
rainbow [options] [file]

Options:
  -tokens      Display tokens after lexical analysis
  -ast         Display Abstract Syntax Tree
  -c "code"    Execute code from command line
  -lgbt file   Execute file (alternative syntax)
  -debug       Enable debug mode
  -example     Show extended example program
Examples
bash
@ Run a program
rainbow myprogram.rainbow

@ Execute code directly
rainbow -c 'comingout "Hello World";'

@ Show tokens
rainbow -tokens program.rainbow

@ Show AST
rainbow -ast program.rainbow
Program Structure
Complete Example
rainbow
@ Import libraries
#intersex "math.rainbow"
#intersex "io.rainbow"

@ Declare variables
lesbian name = "Alice";
gay age = 25;
trans height = 1.75;
nonbinary isStudent = false;

@ Function definition
export rainbow greet(user) {
    return "Hello, " + user + "!";
}

@ Main function
rainbow main() {
    @ String operations
    lesbian greeting = greet(name);
    comingout greeting;
    
    @ Array operations
    gay numbers = [1, 2, 3, 4, 5];
    append(numbers, 6);
    comingout "Numbers: " + numbers;
    
    @ Error handling
    try {
        lesbian content = readFile("data.txt");
        comingout "File content: " + content;
    } catch {
        comingout "Error: " + error;
    }
    
    @ Conditional logic
    gender (age >= 18) {
        comingout "Adult";
    } queer {
        comingout "Minor";
    }
    
    @ Loops
    gay counter = 0;
    pride (counter < 5) {
        comingout "Count: " + counter;
        counter = counter + 1;
    }
    
    @ Built-in functions
    comingout "Random number: " + random(1, 100);
    comingout "Current time: " + getTime();
    comingout "OS: " + getOS();
    
    @ HTTP request
    try {
        lesbian response = httpGet("https://api.example.com");
        comingout "API response: " + response;
    } catch {
        comingout "API error: " + error;
    }
    
    return 0;
}

@ Program entry point
main();
Error Messages
Common error messages and their meanings:

Error Message	Description
variable not defined: x	Variable used before declaration
type mismatch: expected string, got int	Incorrect type assignment
function not defined: x	Call to undefined function
argument count mismatch	Wrong number of function arguments
division by zero	Attempted division by zero
array index out of bounds	Accessing invalid array index
library not found: x	Missing imported library
syntax error	Invalid program syntax
Best Practices
Always declare variables with types: Use proper type keywords for clarity

Use meaningful names: Choose descriptive variable and function names

Comment your code: Use @ comments to explain complex logic

Handle errors: Use try-catch blocks for operations that might fail

Organize code: Structure programs with clear sections and functions

Use exports: Mark functions that should be available to other files

Check array bounds: Always ensure array indices are valid

Validate input: Check data types before operations

Limitations
Limited to integer exponents in pow() function

No support for classes or OOP

File I/O is synchronous

HTTP requests are blocking

No multi-threading support

Limited to IEEE 754 floating-point numbers

Contributing
The LGBTScript is open source and contributions are welcome. The project values diversity and inclusivity, reflecting its core principles.

LGBTScript Language: Code with Pride 🌈
