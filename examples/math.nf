\ nForth Example: Math and Functions

\ Define a function to calculate the square of a number
: SQUARE { n }
    n n MUL INTO result
;

\ Use the function
5 SQUARE INTO five_squared
five_sq PRINT

\ Basic arithmetic with local variables
10 INTO a
20 INTO b
a b ADD INTO sum
sum PRINT
