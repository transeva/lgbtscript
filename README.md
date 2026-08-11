LGBTScript
LGBTScript is a unique, inclusive programming language designed to support and empower the LGBTQ+ community. It combines familiar programming paradigms with a colorful, queer-friendly syntax and a rich set of built‑in functions for social support, mental health, community building, and even web server creation. Written in Go, LGBTScript is both fun and practical.

Features
LGBTQ+‑friendly keywords – lesbian (string), gay (int), trans (float), nonbinary (bool), gender (array), asexual (int).

Intuitive control flow – cis (if), nocis (else if / else), pride (while), homo (for).

Functions – defined with rainbow and optionally exported with export.

Object‑oriented programming – QUEER classes, inheritance (extends), constructors (init), this and super, instantiation with new.

Social‑good built‑ins – crisis support, safe spaces, pride parade info, coming‑out tips, LGBTQ+ history, and much more.

Web server support – create and manage HTTP servers with routing (createServer, startServer, addRoute, etc.).

File system & networking – read/write files, HTTP requests, JSON parsing, hashing (MD5, SHA‑256), regex.

Sandboxing – file access and URL restrictions for safe execution.

Interactive mode – run scripts from files or execute code directly from the command line.

Installation
Clone the repository:

bash
git clone https://github.com/yourusername/lgbtscript.git
cd lgbtscript
Build the binary (requires Go 1.18+):

bash
go build -o lgbtscript
Run:

bash
./lgbtscript [options] [file.rainbow]
Alternatively, download a pre‑built binary from the releases page.

Usage
bash
# Execute a .rainbow file
./lgbtscript myprogram.rainbow

# Run a single command
./lgbtscript -c 'COMINGOUT "Hello, world!"'

# Show tokens or AST for debugging
./lgbtscript -tokens -ast myprogram.rainbow

# Run the extended example
./lgbtscript -example
Command‑line flags
Flag	Description
-tokens	Print lexer tokens
-ast	Print the abstract syntax tree
-c "code"	Execute a code snippet
-lgbt file	Run a .rainbow file (alternative)
-debug	Enable debug output
-example	Show a comprehensive demo
Syntax Guide
Comments
Single‑line comments start with @.

text
@ This is a comment
Data Types
Keyword	Go equivalent	Description
lesbian	string	Text
gay	int	Integer number
trans	float64	Floating‑point number
nonbinary	bool	Boolean (true / false)
gender	[]TypedValue	Array (dynamic)
asexual	int	Integer (synonym for gay)
Variables
Declare a variable with a type:

text
lesbian name = "Alex";
gay age = 25;
trans height = 1.75;
nonbinary isStudent = true;
gender colors = ["red", "green", "blue"];
Assignment (must match declared type):

text
name = "Maria";
age = age + 1;
Control Flow
Conditional: cis / nocis
text
cis (age >= 18) {
    COMINGOUT "You are an adult.";
}
nocis (age >= 13) {
    COMINGOUT "You are a teenager.";
}
nocis {
    COMINGOUT "You are a child.";
}
Loop: pride (while)
text
gay i = 0;
pride (i < 5) {
    COMINGOUT i;
    i = i + 1;
}
Loop: homo (for)
text
homo (gay j = 0; j < 3; j = j + 1) {
    COMINGOUT "Iteration " + j;
}
Functions
Define a function with rainbow:

text
rainbow greet(name) {
    COMINGOUT "Hello, " + name + "!";
}
Return a value:

text
rainbow add(a, b) {
    return a + b;
}
Export a function (makes it accessible from other files):

text
export rainbow multiply(a, b) {
    return a * b;
}
Object‑Oriented Programming (QUEER)
Class Definition
text
QUEER Person {
    lesbian name;
    gay age;
    lesbian gender;

    rainbow init(nameVal, ageVal, genderVal) {
        this.name = nameVal;
        this.age = ageVal;
        this.gender = genderVal;
    }

    rainbow introduce() {
        COMINGOUT "I am " + this.name + ", " + this.age + " years old.";
    }
}
Inheritance
text
QUEER Student extends Person {
    lesbian university;

    rainbow init(nameVal, ageVal, genderVal, universityVal) {
        super.init(nameVal, ageVal, genderVal);
        this.university = universityVal;
    }

    rainbow study() {
        COMINGOUT this.name + " studies at " + this.university;
    }
}
Instantiation
text
gender p = new Person("Alex", 25, "male");
p.introduce();

gender s = new Student("Maria", 20, "female", "MSU");
s.study();
Error Handling
text
try {
    // code that may panic
} catch {
    COMINGOUT "Error: " + error;
}
Include Directives
Import external .rainbow or .rb files:

text
#inclusive "mylibrary.rainbow"
Built‑in Functions
LGBTScript provides a rich standard library. Below is a selection of the most important functions.

File System
Function	Description
readFile(filename)	Returns file content as string.
writeFile(filename, data)	Writes string to a file.
fileExists(filename)	Returns true if file exists.
getDirFiles(directory)	Returns an array of filenames in the directory.
String & Array Helpers
Function	Description
split(text, delim)	Splits string into array.
replace(text, old, new)	Replaces all occurrences.
trim(text)	Removes leading/trailing whitespace.
length(val)	Length of string or array.
toUpper(text)	Uppercase conversion.
toLower(text)	Lowercase conversion.
append(arr, elem)	Adds element to array.
remove(arr, index)	Removes element at index.
Math & Random
Function	Description
random(min, max)	Random integer in range [min, max].
max(a, b, ...)	Maximum of numbers.
min(a, b, ...)	Minimum of numbers.
sqrt(x)	Square root.
pow(base, exp)	Power (integer exponent only).
Time & System
Function	Description
getTime()	Current date/time string.
getYear()	Current year as integer.
getMonth()	Current month as integer (1‑12).
getOS()	Operating system name.
getHostname()	Machine hostname.
getArgs()	Command‑line arguments array.
hasFlag(flag)	Check if a flag is present.
Networking & JSON
Function	Description
httpGet(url)	Perform GET request, returns body.
httpPost(url, data)	POST JSON data, returns response.
jsonParse(jsonString)	Parse JSON into LGBTScript value.
md5(text)	Return MD5 hash as hex string.
sha256(text)	Return SHA‑256 hash.
regexFind(pattern, text)	Find all matches, return array.
regexReplace(pattern, text, repl)	Replace with regex.
Web Server (Experimental)
Function	Description
createServer(name, port)	Creates a server instance.
startServer(name)	Starts the server.
stopServer(name)	Stops the server.
addRoute(server, method, path, handler)	Adds a route (handler is a rainbow function).
getServerStatus(name)	Returns server status info.
listServers()	Lists all created servers.
Social & LGBTQ+ Support
Function	Description
findSafeSpace(place, city, radius)	Find LGBTQ+‑friendly places.
getCrisisSupport(region, type)	Crisis hotlines and resources.
getLGBTQLaws(country, category)	Legal information per country.
getDailyAffirmation(theme)	Motivational affirmation.
moodCheck(moods, suggestResources)	Mental health check‑in.
guidedBreathing(minutes, theme)	Breathing exercise for stress relief.
defineTerm(term, language)	Definition of LGBTQ+ terms.
lgbtHistoryQuiz(difficulty)	Quiz about LGBTQ+ history.
getDailyFact(category, region)	Random LGBTQ+ fact.
getHRTInfo(country, hrtType)	Hormone therapy information.
findLGBTDoctor(specialty, city)	Find friendly doctors.
getDocumentChangeGuide(country, document)	Guide for changing legal documents.
getLGBTQEvents(days, type)	Upcoming events.
createLGBTQGroup(name, meetingType)	Start a support group.
findVolunteerOpportunity(org, skills)	Volunteer positions.
getLGBTQBook(genre)	Book recommendations.
getLGBTQPlaylist(mood)	Music playlist by mood.
getLGBTQMovies(genre)	Movie recommendations.
getPrideParadeInfo(city, year)	Pride parade details.
getComingOutTips(audience)	Advice for coming out.
getTransHealthcare(country)	Trans‑specific healthcare.
findLGBTQShelter(location)	Shelters and safe housing.
getIntersexResources(country)	Intersex support resources.
getNonbinaryGuide()	Guide for non‑binary individuals.
findLGBTQTherapist(specialty, location)	Find LGBTQ+‑friendly therapists.
getAsylumInfo(country)	Asylum and refugee information.
getAsexualResources()	Resources for asexual people.
getPolyamoryGuide()	Guide to ethical non‑monogamy.
getGenderAffirmingCare(country)	Gender‑affirming care info.
findLGBTQCommunity(interest)	Find communities by interest.
getLGBTQHistory(era)	Historical overview.
getLGBTQParenting()	Parenting resources for LGBTQ+ families.
getConversionTherapyHelp()	Help for survivors of conversion therapy.
getLGBTQHousing(location)	LGBTQ+‑friendly housing.
getQueerArt(medium)	Queer art and artists.
getLGBTQFriendlyCities(criteria)	List of LGBTQ+‑friendly cities.
Special Statements
orientation; – runs an interactive sexual orientation psychological test (or demo in file mode).

help "country"; – displays LGBTQ+ support organisations for that country.

Example
A full example showcasing most features:

text
@ Extended LGBTScript example

QUEER Person {
    lesbian name;
    gay age;
    lesbian gender;

    rainbow init(nameVal, ageVal, genderVal) {
        this.name = nameVal;
        this.age = ageVal;
        this.gender = genderVal;
        COMINGOUT "Person created: " + this.name;
    }

    rainbow introduce() {
        COMINGOUT "Hi, I'm " + this.name + ", " + this.age + " years old.";
    }
}

QUEER Student extends Person {
    lesbian university;

    rainbow init(nameVal, ageVal, genderVal, universityVal) {
        super.init(nameVal, ageVal, genderVal);
        this.university = universityVal;
    }

    rainbow study() {
        COMINGOUT this.name + " studies at " + this.university;
    }
}

rainbow main() {
    COMINGOUT "🌈 Welcome to LGBTScript!";

    gay i = 0;
    pride (i < 5) {
        COMINGOUT "pride loop #" + i;
        i = i + 1;
    }

    homo (gay j = 0; j < 3; j = j + 1) {
        COMINGOUT "homo loop #" + j;
    }

    gender p = new Person("Alex", 25, "male");
    p.introduce();

    gender s = new Student("Maria", 20, "female", "MSU");
    s.introduce();
    s.study();

    COMINGOUT "Today is " + getTime();
    COMINGOUT "Pride parade info: " + getPrideParadeInfo("Moscow", 2026);
    COMINGOUT "Coming out tips: " + getComingOutTips("parents");

    gay server = createServer("myapi", 8080);
    COMINGOUT server;

    addRoute("myapi", "GET", "/hello", rainbow (query, body) {
        return "Hello from LGBTScript!";
    });

    startServer("myapi");

    COMINGOUT "All done!";
}

main();
License
Copyright © 2026 LGBTScript contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Contributing
Contributions are welcome! Please open issues and pull requests on GitHub.
We strive to maintain a friendly, inclusive community.

Acknowledgments
Inspired by the LGBTQ+ community and built with love.
Special thanks to all contributors and supporters.

Happy coding! 🌈

