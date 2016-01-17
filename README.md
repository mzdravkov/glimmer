# Glimmer
A Go tool that visualizes the communication between goroutines

Glimmer works by modifying the AST ([Abstract Syntax Tree](https://en.wikipedia.org/wiki/Abstract_syntax_tree)) of the program that you provide it with. It then compiles this modified copy of your program (it won't change the actual code, don't worry) and runs it. The modifications that Glimmer did in the first step will send all the messages that are being send by channels, that are in a tracked functions, to a front-end that will display this messages nicely.

# Installation

```go get github.com/mzdravkov/glimmer```

# How to use it
Glimmer is pretty simple and straight forward. All you need to do is to annotate the functions that you want to examine with this special comment:

```// glimmer```

Note: it should be written exactly as it appears here, this means that if you add an extra space (or no space at all) between the comment token and the word ```glimmer``` it won't work.

Once you've finished annotating, you just run the glimmer command-line tool and provide it with the path to the program:

```glimmer on /path/to/the/program```

Voila!

# Front-end
When Glimmer runs it starts a web server listening on port 9610. Just open ```localhost:9610``` to start watching the visualization of your goroutines communicating. When Glimmer starts your program it will block the execution until you start it from the front end. This is done in order to not miss the beginning of your program's visualization. This behaviour can be altered by providing a ```--no-wait``` flag.

# Flags
    --no-wait            disables the initial block that stops the program until you start it from the front-end
    --port 10000         changes the default port of the web server that powers the front-end
    --log /path/to/file  logs the JSON that is used for communication between the back-end and the front-end
    --no-front-end       disables the server that powers the front-end
    --delay 1000         sets the default delay for messages in milliseconds
    --sequentialize      sets the default mode to sequential
    --examine-all        visualizes all goroutines, not just the ones that are annotated


# State
Still, ```Glimmer``` is in a very early state - totally not ready for work

# License
Apache License Version 2.0, January 2004


See the LICENSE document in the project tree root for details.

