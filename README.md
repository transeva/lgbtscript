
> *A programming language where every 💖 matters*

**Version 2.0** · **Pride‑first** · **Inclusive syntax**

---

## 📖 Introduction

**LGBTScript** is an esoteric programming language crafted with love and sarcasm.  
Its syntax is inspired by the values of the LGBTQ+ community: *diversity, respect, self‑expression*.  
Every construct is a little celebration, and code is written for people, not just machines.

LGBTScript doesn't aim to be practical – it aims to bring a smile and remind us that programming can be **kind** and **colourful**.

> 🏳️‍🌈 **Philosophy:** *“Code is art. Every developer is unique, just like every colour in the rainbow.”*

---

## 🧬 Data Types

All values are **emotions** and **entities**.

| Type        | Description                    | Example                |
|-------------|--------------------------------|------------------------|
| `Pride`     | Integer (pride)                | `42`                   |
| `Rainbow`   | String (colourful text)        | `"hello 🌈"`           |
| `Queer`     | Boolean (true / false)         | `true` / `false`       |
| `NonBinary` | List (array)                   | `[1, 2, 3]`            |
| `Trans`     | Dictionary (key‑value)         | `{name: "Alex"}`       |
| `Fluid`     | Floating‑point number          | `3.14`                 |

---

## ✍️ Syntax

### Variables

Use the keyword `let` (as in “let yourself be”).

```javascript
let myName = "Alex"   // Rainbow string
let age = 25          // Pride number
let isAwesome = true  // Queer boolean
Conditionals – pride / queer
Instead of if / else we write pride / queer.

javascript
pride (age > 18) {
  print("Adult")
} queer {
  print("Still young")
}
Loops – rainbow
The rainbow loop iterates over all colours of the rainbow (or list items).

javascript
let colors = ["red", "orange", "yellow"]
rainbow (color in colors) {
  print(color)
}
For an infinite loop – rainbow infinity.

Functions – sparkle
Define functions with sparkle (a spark).

javascript
sparkle greet(name) {
  return "Hello, " + name + " ✨"
}

let message = greet("Sam")
🧩 Standard Library – Complete Examples
print( ... )
Outputs any value to the console.

javascript
print("Hello, world!")                // → Hello, world!
print(42)                             // → 42
print([1, 2, 3])                      // → [1, 2, 3]
print({name: "Alex", age: 25})        // → {name: "Alex", age: 25}
love( a, b )
Returns the sum of a and b (because love adds up).

javascript
let result = love(5, 3)               // → 8
print(result)

// Works with Pride, Fluid, even Rainbow strings (concatenation)
let greeting = love("Hello, ", "world!")  // → "Hello, world!"
hug( a, b )
Returns the product of a and b (a warm embrace multiplies joy).

javascript
let area = hug(4, 5)                  // → 20
print("Area: " + area)

// Also works with strings (repetition)
let stars = hug("*", 5)               // → "*****"
support( condition, thenValue, elseValue )
Ternary operator – returns thenValue if condition is true, otherwise elseValue.

javascript
let age = 19
let status = support(age >= 18, "adult", "minor")
print(status)                         // → "adult"

// Can be nested
let x = 7
let parity = support(x % 2 == 0, "even", "odd")
prideLen( value )
Returns the length of a Rainbow (string) or NonBinary (list).
For other types, it returns a prideful 1 (because you're always enough).

javascript
let text = "🌈rainbow🌈"
let len = prideLen(text)              // → 10 (counts Unicode graphemes)

let list = [1, 2, 3, 4]
print(prideLen(list))                 // → 4

let number = 42
print(prideLen(number))               // → 1  (everything is beautiful)
rainbowJoin( list, separator )
Joins a NonBinary list into a Rainbow string, using the given separator.

javascript
let fruits = ["apple", "banana", "cherry"]
let joined = rainbowJoin(fruits, ", ")
print(joined)                         // → "apple, banana, cherry"

// Also works with numbers and other types
let mixed = [1, 2, 3]
let result = rainbowJoin(mixed, " - ") // → "1 - 2 - 3"
🎯 Complete Program Examples
FizzBuzz (Pride edition)
javascript
sparkle fizzbuzz(n) {
  let output = ""
  pride (n % 15 == 0) {
    output = "FizzBuzz"
  } queer pride (n % 3 == 0) {
    output = "Fizz"
  } queer pride (n % 5 == 0) {
    output = "Buzz"
  } queer {
    output = str(n)   // convert Pride to Rainbow
  }
  return output
}

rainbow (i in range(1, 16)) {
  print(fizzbuzz(i))
}
// Output: 1, 2, Fizz, 4, Buzz, Fizz, 7, 8, Fizz, Buzz, 11, Fizz, 13, 14, FizzBuzz
Working with Dictionaries (Trans)
javascript
let person = {name: "Jordan", age: 30, city: "Berlin"}

// Access values
print(person.name)      // → "Jordan"
print(person["age"])    // → 30

// Update
person.city = "Amsterdam"
print(person.city)      // → "Amsterdam"
List Operations
javascript
let nums = [1, 2, 3, 4, 5]
let squares = []

rainbow (n in nums) {
  let sq = hug(n, n)     // square using hug
  squares = rainbowJoin([squares, sq], ", ")   // (simulated push)
}

print(squares)           // → "1, 4, 9, 16, 25"
🧪 Error Handling – try / embrace
Use try and embrace (instead of catch) to handle errors.

javascript
try {
  let risky = love(10, "not a number")   // type mismatch
} embrace (error) {
  print("It's okay, we embrace you: " + error)
}
✨ Fun fact: All errors are called “misgender” and are printed in purple.

📦 Modules and Imports – ally
Import other files with ally (ally).

javascript
ally "math.lgbt"
ally "utils.lgbt" as utils

let result = utils.someFunction(5)
🤝 Community & Contributions
LGBTScript lives because of its community. We welcome pull requests, bug reports, and new ideas.
The golden rule: be kind to code and to people.

Official repository: github.com/lgbt-script/lgbt-script

📄 License
MIT · Made with ❤️ and rainbows.
