# Borno Programming Language

Borno is a **Bangla-based** programming language that allows developers to write code using **Bangla keywords** and identifiers. Its goal is to provide a familiar programming experience to native Bangla speakers, while still supporting typical programming constructs like variables, functions, arrays, objects, loops, and more.

---

## Table of Contents

- [Borno Programming Language](#borno-programming-language)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Installation](#installation)
  - [How Borno Works](#how-borno-works)
  - [Usage](#usage)
  - [Core Grammar](#core-grammar)
  - [Keywords \& Reserved Words](#keywords--reserved-words)
  - [Examples](#examples)
    - [Native Library Demo](#native-library-demo)
    - [Array \& Object Demo](#array--object-demo)
    - [Control Flow Demo](#control-flow-demo)
    - [Closures \& Functions](#closures--functions)

---

## Features

- **Bangla Keywords**: Write if-statements, loops, and function declarations in Bangla.
- **Arrays & Objects**: Use `[ ]` and `{ }` for arrays and objects, respectively.
- **Functions**: Define custom functions with parameters, closures, and return statements.
- **Built-In Functions**: Access native functions like input, array manipulation (append, remove), math utilities (sqrt, abs, sin, etc.).
- **Bangla Digits**: Parse and convert Bangla digits (০, ১, ২, ৩, ...) to ASCII under the hood.

---

## Installation

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/ah-naf/borno.git
   cd borno
   ```

2. **Build** (assuming you have Go installed):

   ```bash
   go build -o borno
   ```

   This produces an executable named `borno`.

---

## How Borno Works

Borno follows a **three-phase** process:

1. **Lexical Analysis (Lexer)**  
   The source code (in `.bn` files) is scanned character-by-character to produce **tokens**. For instance, `ধরি`, `ফাংশন`, `যদি`, etc., are recognized as **Bangla keywords**, while identifiers and operators are tokenized accordingly.

2. **Parsing (Parser)**  
   The tokens are read according to Borno’s **grammar rules**, building an **Abstract Syntax Tree** (AST). This phase checks syntax (e.g., matching parentheses, valid expressions).

3. **Interpretation**  
   The AST is **walked** by an **interpreter**, which executes each statement and expression. There are native functions (e.g., `ইনপুট`, `লেন`, `এড`) that integrate Go-based capabilities (like reading from stdin or calculating math functions).

When you run a Borno script, these steps occur behind the scenes. Any errors (syntax or runtime) are displayed in Bangla or Banglish messages.

---

## Usage

Once you’ve built the `borno` executable:

1. **Run a Borno Script**:

   ```bash
   ./borno my_script.bn
   ```

2. **Interactive Mode (REPL)**:
   If you run `./borno` with no file arguments, you can type code line by line. This is useful for quick tests or demos.

3. **Examples**:
   Check out the `examples/` directory (or see below) for `.bn` files demonstrating language features.

**File Extension**: We recommend using `.bn` (short for “Borno”) for all source files.

---

## Core Grammar

Below is a **simplified** version of Borno’s grammar:

```
program        → declaration* EOF ;

declaration    → funDecl
               | varDecl
               | statement ;

funDecl        → "ফাংশন" function ;
function       → IDENTIFIER "(" parameters? ")" block ;
parameters     → IDENTIFIER ( "," IDENTIFIER )* ;

varDecl        → "ধরি" variable ( "," variable )* ";" ;
variable       → IDENTIFIER ( "=" expression)? ;

statement      → exprStmt
               | ifStmt
               | whileStmt
               | forStmt
               | printStmt
               | block
               | breakStmt
               | continueStmt
               | returnStmt ;

ifStmt         → "যদি" "(" expression ")" statement ( "নাহয়" statement )? ;
whileStmt      → "যতক্ষণ" "(" expression ")" statement ;
forStmt        → "ফর" "(" ( varDecl | exprStmt | ";" ) expression? ";" expression? ")" statement ;
exprStmt       → expression ";" ;
printStmt      → "দেখাও" expression ";" ;
block          → "{" declaration* "}" ;
breakStmt      → "থামো" ";" ;
continueStmt   → "চালিয়ে_যাও" ";" ;
returnStmt     → "ফেরত" expression? ";" ;

expression     → assignment ;
assignment     → IDENTIFIER "=" assignment | logic_or ;

logic_or       → logic_and ( ( "বা" | "||" ) logic_and )* ;
logic_and      → equality ( ( "এবং" | "&&" ) equality )* ;
equality       → comparison ( ( "!=" | "==" ) comparison )* ;
comparison     → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
term           → factor ( ( "-" | "+" ) factor )* ;
factor         → power ( ( "/" | "*" | "%" ) power )* ;
power          → unary ( ( "**" ) unary )* ;
unary          → ( "!" | "-" | "~" ) unary | primary ;
primary        → NUMBER | STRING | "সত্য" | "মিথ্যা" | "nil" | "(" expression ")" | IDENTIFIER
               | arrayLiteral | objectLiteral ;

arrayLiteral   → "[" ( expression ( "," expression )* )? "]" ;
objectLiteral  → "{" ( property ( "," property )* )? "}" ;
property       → IDENTIFIER ":" expression ;
```

---

## Keywords & Reserved Words

Here are the **Bangla keywords** Borno uses:

| Keyword         | Description                |
|-----------------|----------------------------|
| `ফাংশন`         | Declares a function.      |
| `ধরি`           | Declares a variable.      |
| `ফর`            | For-loop.                 |
| `যদি`           | If-statement.             |
| `নাহয়`          | Else-statement.           |
| `যতক্ষণ`       | While-loop.               |
| `সত্য`          | Boolean true.             |
| `মিথ্যা`        | Boolean false.            |
| `দেখাও`         | Print statement.          |
| `ফেরত`          | Return from function.     |
| `থামো`          | Break from loop.          |
| `চালিয়ে_যাও`    | Continue loop.            |
| `এবং`           | Logical AND (&&).         |
| `বা`            | Logical OR (&#124;&#124;).|

Reserved identifiers like `ক্লক`, `ইনপুট`, `এড`, `রিমুভ`, etc., are bound to **native functions** in the global environment.

---

## Examples

Below are a few snippet examples to illustrate various features of Borno.

### Native Library Demo

Save this code as **`native_library_demo.bn`** and run `borno native_library_demo.bn`:

```none
// 1) ক্লক (clock) 
//    Shows the current timestamp in seconds.
দেখাও "বর্তমান সময় (সেকেন্ডে): " + ক্লক();

// 2) ইনপুট (input) 
//    Uncomment these lines to test user input interactively.
// দেখাও "কিছু লিখুনঃ"
ধরি প্রবেশ = ইনপুট("আপনার লেখা: ");
দেখাও "আপনি লিখেছেনঃ " + প্রবেশ;

// 3) লেন (len)
//    Returns the length of an array.
ধরি তালিকা = [১০, ২০, ৩০];
দেখাও লেন(তালিকা);

// 4) এড (append)
//    Appends one or more elements to the array.
তালিকা = এড(তালিকা, ৪০);
দেখাও তালিকা;

// 5) রিমুভ (remove)
//    Removes an element from the array at a given index.
তালিকা = রিমুভ(তালিকা, ১);
দেখাও তালিকা;

// 6) কি_রিমুভ (delete)
//    Deletes a property from an object by key.
ধরি বস্তু = {
    নাম: "বর্ণলিপি",
    ধরন: "ডেমো",
    মান: ৫
};
দেখাও বস্তু;
কি_রিমুভ(বস্তু, "মান");
দেখাও বস্তু;

// 7) অব্জেক্ট_কি (keys) এবং অব্জেক্ট_মান (values)
//    Gets arrays of keys or values from an object.
দেখাও অব্জেক্ট_কি(বস্তু);
দেখাও অব্জেক্ট_মান(বস্তু);

// 8) পরমমান (abs)
//    Returns the absolute value of a number.
দেখাও "পরমমান(-১২.৫): " + পরমমান(-১২.৫);

// 9) বর্গমূল (sqrt)
//    Computes the square root of a number.
দেখাও "বর্গমূল(১৬): " + বর্গমূল(১৬);

// 10) ঘাত (pow)
//     Raises the first number to the power of the second.
দেখাও "ঘাত(২, ৮): " + ঘাত(২, ৮);

// 11) সাইন (sin), কসাইন (cos), ট্যান (tan)
//     Common trigonometric functions in radians.
দেখাও "সাইন(৩.১৪/২) => " + সাইন(৩.১৪ / ২);
দেখাও "কসাইন(০) => " + কসাইন(০);
দেখাও "ট্যান(৩.১৪/৪) => " + ট্যান(৩.১৪ / ৪);

// 12) সর্বনিম্ন (min), সর্বোচ্চ (max)
//     Returns the smallest/largest value among the given numbers.
দেখাও "সর্বনিম্ন(১০, ৫, -৩, ৮) => " + সর্বনিম্ন(১০, ৫, -৩, ৮);
দেখাও "সর্বোচ্চ(১০, ৫, -৩, ৮) => " + সর্বোচ্চ(১০, ৫, -৩, ৮);

// 13) রাউন্ড (round)
//     Rounds a floating-point number to the nearest integer.
দেখাও "রাউন্ড(৩.৬৭) => " + রাউন্ড(৩.৬৭);

```

And so on. This snippet demonstrates user input, array functions, object manipulation, etc.

---

### Array & Object Demo

```none
ফাংশন testArrayAndObject() {
    ধরি arr = [10, 20, 30];
    দেখাও "First element of arr: " + arr[0];
    arr[2] = 300;
    দেখাও "Modified third element of arr: " + arr[2];

    ধরি obj = {
        name: "Borno Language",
        count: 1
    };
    দেখাও "Object name property: " + obj.name;
    obj.count = obj.count + 1;
    দেখাও "Updated count property: " + obj.count;

    ফেরত obj;
}

ধরি result = testArrayAndObject();
দেখাও "Returned object count: " + result.count;
```

---

### Control Flow Demo

```none
// If-Else
ধরি x = 15;
যদি (x > 10) {
    দেখাও "x is greater than 10";
} নাহয় {
    দেখাও "x is 10 or less";
}

// For loop with break/continue
ফর (ধরি i = 0; i < 5; i = i + 1) {
    যদি (i == 2) {
        দেখাও "Skipping i = 2";
        চালিয়ে_যাও;
    }
    যদি (i == 4) {
        দেখাও "Breaking at i = 4";
        থামো;
    }
    দেখাও i;
}

// While loop
ধরি count = 0;
যতক্ষণ (count < 3) {
    দেখাও "Count = " + count;
    count = count + 1;
}
```

---

### Closures & Functions

```none
ফাংশন createCounter() {
    ধরি counter = 0;

    ফাংশন increment() {
        counter = counter + 1;
        দেখাও "Counter = " + counter;
    }

    ফেরত increment;
}

ধরি counter1 = createCounter();
counter1(); // Counter = 1
counter1(); // Counter = 2
counter1(); // Counter = 3
```

---
