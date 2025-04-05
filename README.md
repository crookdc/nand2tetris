# nand2tetris
Toy project following the book *The Elements of Computing Systems* and its accompanying project 'Hack', written in Go.

## HDL
This project comes with its own implementation of a Hardware Definition Language (HDL). The syntax of this language is 
rather simple, perhaps a bit too simple. I might consider adding some syntactic sugar to it in the future to spice up 
the projects HDL files and make them look a bit cleaner.

### Language Definition
Below is a very minimal example of a simple chip definition, namely, the *not gate*. Most of the features of the HDL 
language can be found within this simple example so let's break it down.

```
chip not (in: 1) -> (1) {
	out nand(in: [in.0, 1])
}
```

The chip header serves as a declaration of the interface for the chip. It includes a name, a list of named input pins 
and a list of anonymous (unnamed) output pins. In the example above we can deduce that the chip name is `not` and that 
the input pins are defined as `(in: 1)`, meaning that there is a single input pin called `in` that accepts a single 
signal. Alternatively, you could have more input pins and end up with something like `(selector: 1, a: 16, b: 16)` where 
there are three input pins, `selector`, `a` and `b`, where `a` and `b` accept up to 16 signals. We can also see from the 
example above that this chip provides a single output pin which will emit a single signal. This definition is found 
immediately after the arrow (`->`) and is simply a comma-separated list of positive integers. 

What follows next in our example is the chip body, which defines how the chip defines the values of its outputs as a 
function of its inputs. In our example we define it as the result of passing the values `in.0` (the first and only 
signal of `in`) and a constant 1 to a NAND gate. The `out` keyword is used to define the value of an output pin and each 
occurrence of the `out` keyword directly maps to the respective spot in the chip header.

It is also possible to wire a chips output to itself using a special keyword, namely *feedback*. Once feedback is 
requested then all the output pins of the current chip will be wired into the assigned names. The names are defined 
using the `set` keyword which can be used to create any pin that is not directly part of the input interface of the 
chip itself. An example of using the feedback pins can be viewed below.

```
chip not_feedback (in: 1) -> (1, 1) {
    set _, previous = feedback()
    out previous
    out not(in: in)
}
```

Here you can see the `set` and `feedback` keywords in action, and by naming a pin `_` we can avoid its creation
altogether if it is not needed. You might also notice that we did not index the `in` parameter we passed to the `not` 
chip, this is because both pins are of the same size and can therefore be treated as their own units.

Finally, to modularize the implementation of a great many chips this HDL allows you to import other HDL files into the 
current one to use the chips defined there. However, circular dependencies are not being checked for so please handle 
with care. A simple example of importing another file is shown below.

```
use "not.hdl"

chip or (in: 2) -> (1) {
    out nand(in: [not(in: in.0), not(in: in.1)])
}
```

It's that simple, just provide a relative path to the file you wish to use after the `use` keyword. You might have 
noticed that `nand` is used within the chip body but not imported anywhere. That is totally valid since the `nand` and 
`dff` gates are builtin chips that can be used without any imports.