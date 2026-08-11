# 🌈 LGBTScript

**LGBTScript** is an inclusive programming language with LGBTQ+ community support, built‑in social functions, an embedded web server, and OOP based on `QUEER` classes.

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

---

## Table of Contents

- [Introduction](#introduction)
- [Installation & Running](#installation--running)
- [Syntax & Keywords](#syntax--keywords)
- [Data Types](#data-types)
- [Variables & Declarations](#variables--declarations)
- [Control Flow](#control-flow)
- [Functions (`rainbow`)](#functions-rainbow)
- [OOP: QUEER Classes](#oop-queer-classes)
- [Arrays](#arrays)
- [Built‑in Functions](#builtin-functions)
- [Social Functions](#social-functions)
- [Server Functions](#server-functions)
- [Error Handling](#error-handling)
- [The `#inclusive` Directive](#the-inclusive-directive)
- [Code Examples](#code-examples)
- [License](#license)

---

## Introduction

LGBTScript is an interpreted language created to support the LGBTQ+ community. It combines a friendly syntax with powerful features: file I/O, HTTP requests, cryptography, web server creation, social functions (help, resources, psychological support), and full OOP.

---

## Installation & Running

Compile the source code (`lgbt.go`) with Go:

```bash
go build -o lgbt lgbt.go
Run:

bash
./lgbt [flag] [file.rainbow]
Flags:

Flag	Description
-c "code"	Execute code from the command line
-tokens	Print tokens (lexical analysis)
-ast	Print the AST (syntax tree)
-debug	Enable debug mode
-example	Show an extended example
-lgbt file	Execute a code file
You can also specify a file as a positional argument.

Syntax & Keywords
Keywords (case‑insensitive):

lesbian – string type

gay / asexual – integer type

trans – floating‑point type

nonbinary – boolean type

gender – array type

comingout – print output

cis – conditional (if)

nocis – else if / else

pride – while loop

homo – for loop

rainbow – function declaration

return – return from function

help – display information about LGBTQ+ organisations

orientation – run an orientation test

try / catch – exception handling

export – export a function (for libraries)

queer – class declaration

new – instantiate a class

this / super – references to current/parent object

extends – inheritance

Comments start with @.

Data Types
Keyword	Type	Example
lesbian	string	"Hello"
gay / asexual	integer	42
trans	float	3.14
nonbinary	boolean	true / false
gender	array	[1, 2, 3]
Variables & Declarations
rainbow
lesbian name = "Alex";
gay age = 25;
trans height = 1.75;
nonbinary isStudent = true;
gender colors = ["red", "green", "blue"];

name = "Maria";
age = age + 1;
If no initial value is given, a default is used (empty string, 0, 0.0, false, empty array).

Control Flow
cis (if)
rainbow
cis (age >= 18) {
    comingout "Adult";
} nocis (age > 12) {
    comingout "Teenager";
} nocis {
    comingout "Child";
}
pride (while)
rainbow
gay i = 0;
pride (i < 5) {
    comingout i;
    i = i + 1;
}
homo (for)
rainbow
homo (gay j = 0; j < 3; j = j + 1) {
    comingout "j = " + j;
}
Functions (rainbow)
rainbow
rainbow sum(a, b) {
    return a + b;
}
rainbow main() {
    comingout sum(5, 7); // 12
}
Functions can be exported with export.

OOP: QUEER Classes
rainbow
queer Person {
    lesbian name;
    gay age;
    
    rainbow init(nameVal, ageVal) {
        this.name = nameVal;
        this.age = ageVal;
    }
    
    rainbow introduce() {
        comingout "Hello, I am " + this.name;
    }
}

queer Student extends Person {
    lesbian university;
    
    rainbow init(nameVal, ageVal, univ) {
        super.init(nameVal, ageVal);
        this.university = univ;
    }
}

gender p = new Person("Alex", 25);
p.introduce();
Arrays
rainbow
gender arr = [1, 2, 3];
comingout arr[0]; // 1
arr[1] = 99;
Built‑in Functions
File I/O: readFile, writeFile, fileExists, getDirFiles

Strings: split, replace, trim, length, toUpper, toLower

Arrays: append, remove

Math: random, max, min, sqrt, pow

Time: getTime, getYear, getMonth

System: getOS, getHostname, getArgs, hasFlag

HTTP: httpGet, httpPost

JSON: jsonParse

Cryptography: md5, sha256

Regex: regexFind, regexReplace

Other: sleep, sendEmail (demo), flag (returns the rainbow colours).

Social Functions
findSafeSpace(place, city, radius)

getCrisisSupport(region, type)

getLGBTQLaws(country, category)

getDailyAffirmation(theme)

moodCheck(moods, suggestResources)

guidedBreathing(minutes, theme)

defineTerm(term, language)

lgbtHistoryQuiz(difficulty)

getDailyFact(category, region)

getHRTInfo(country, hrtType)

findLGBTDoctor(specialty, city)

getDocumentChangeGuide(country, document)

getLGBTQEvents(days, type)

createLGBTQGroup(name, meetingType)

findVolunteerOpportunity(organization, skills)

getLGBTQBook(genre)

getLGBTQPlaylist(mood)

getLGBTQMovies(genre)

getPrideParadeInfo(city, year)

getComingOutTips(audience)

getTransHealthcare(country)

findLGBTQShelter(location)

getIntersexResources(country)

getNonbinaryGuide()

findLGBTQTherapist(specialty, location)

getAsylumInfo(country)

getAsexualResources()

getPolyamoryGuide()

getGenderAffirmingCare(country)

findLGBTQCommunity(interest)

getLGBTQHistory(era)

getLGBTQParenting()

getConversionTherapyHelp()

getLGBTQHousing(location)

getQueerArt(medium)

getLGBTQFriendlyCities(criteria)

Server Functions
rainbow
createServer("myapi", 8080);
addRoute("myapi", "GET", "/hello", rainbow(query, body) {
    return "Hello, LGBTQ+!";
});
startServer("myapi");
Available functions:

createServer(name, port)

startServer(name)

stopServer(name)

addRoute(server, method, path, handler)

getServerStatus(name)

listServers()

The handler receives query (array of key‑value pairs) and body (string).

Error Handling
rainbow
try {
    // code that might fail
} catch {
    comingout "Error: " + error;
}
The #inclusive Directive
rainbow
#inclusive "math.rainbow"
Searches in the current folder, the file's folder, and libs/, libraries/.

Code Examples
A full example with loops, classes, and a server:

rainbow
@ Example LGBTScript
queer Person {
    lesbian name;
    gay age;
    rainbow init(n, a) { this.name = n; this.age = a; }
    rainbow greet() { comingout "Hello, I am " + this.name; }
}

rainbow main() {
    gender p = new Person("Alex", 25);
    p.greet();
    
    comingout createServer("api", 8080);
    comingout addRoute("api", "GET", "/hello", 
        rainbow(query, body) { return "Hello, LGBTQ+!"; }
    );
    comingout startServer("api");
}
main();
License
Copyright 2025 LGBTScript Team

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Both versions are now in **English** and include all the required sections, examples, and the Apache 2.0 license. You can use the HTML as a standalone documentation page and the `README.md` on GitHub.
