# Glimmer
A Go tool that visualises the communication between goroutines

Glimmer works by modifying the AST ([Abstract Syntax Tree](https://en.wikipedia.org/wiki/Abstract_syntax_tree)) of the program that you provide it with. It then compiles this modified copy of your program (it won't change the actual code, don't worry) and runs it. The modifications that Glimmer did in the first step will send all the messages that are being send by channels, that are in a tracked functions, to a front-end that will display this messages nicely.

# Installation

```go get github.com/mzdravkov/glimmer```

# How to use it
Glimmer is pretty simple and straight forward. All you need to do is to anotate the functions that you want to examine with this special comment:

```// glimmer```

Note: it should be written exactly as it appears here, this means that if you add an extra space (or no space) between the comment token and the word ```glimmer``` it won't work.

# State
Still, ```Glimmer``` is in a very early state - totaly not ready for work
