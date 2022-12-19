# FIVE WORDS CHALLENGE
This challenge has been taken from [Matt Parker's YouTube channel](https://www.youtube.com/watch?v=_-AfhLQfb6w).
The aim is to find all combinations of five five-letter words that collectively have 25 unique
letters.

For comparability, Matt has started with an official word list from Wordle.

Matt found all 831 solutions to this puzzle with his "unoptimised" code that took about 32 days to run.
However, a keen viewer quickly managed to find a solution that ran in just 15 minutes,
which then launched a wave of submissions that incorporated better algorithms and faster programming languages.


I found a solution that runs in ~2.8 seconds on my "averagely good work laptop" which is about 3 years old.

### My approach
I was inspired mainly by Matt's [follow-up video](https://www.youtube.com/watch?v=c33AZBnRHks&t=816s). In particular:
 - removing anagrams before solving
 - using a bit-wise representation of words
 - using graph theory
 - ordering alphabet by frequency of letters 

I found the idea to learn more about bit-wise approach particularly interesting and inspired me to have a go.
I also thought that this would lend itself very well to the Numpy Python library allowing for a fast solution in a traditionally "slow" language.

Through my solution, I also converged on utilising techniques employed by other solutions such as aggressive pruning.

### Breakdown of approach

#### Numpy
In general in this approach, we want to use as many vectorized operations as possible,
so this favours a breadth-first (with aggressive pruning) rather a more iterative depth-first approach.

#### Filter on word-set to solve for
 - filter on five-letter words that contain five-unique letters
 - remove anagrams (these can be looked up later)

#### Bit-wise representation
 - covert characters to 26 bit-wise representation such that:
   - A = 00000000000000000000000001
   - B = 00000000000000000000000010
   - C = 00000000000000000000000100
   - ...
   - Z = 10000000000000000000000000
   - This means the alphabet can also be considered as A=2^0, B=2^1, C=2^2...Z=2^26
   - As every word contains unique letters, and everyone combination of words contains unique letters,
     - we can easily determine which letters are contained in our word or combination of words
       - e.g. BADGE 
         - = 00000000000000000001011011
         - = 2^0 + 2^1 + 2^3 + 2^4 + 2^6 = 91

#### Reverse frequency order of alphabet

 - The next step is to find the frequency that letters occur in the overall word-set.
   - This allows us to reduce massively the starting number of nodes and therefore the number of combinations to solve for.
   - For our word-set the frequency of letters from low to high is:
     - 'q', 'x', 'j', 'z', 'v', 'f', 'w', 'b', 'k', 'g', 'p', 'm', 'h', 'd', 'c', 'y', 't', 'l', 'n', 'u', 'r', 'o', 'i', 's', 'e', 'a'
   - As our five-word solution must contain 25 letters, we can say that the first word must contain a 'q' or an 'x'
     - thus reducing significantly the starting combinations
   - The next word in each solution must contain the next lowest (or second lowest), unused letter
     - As we are only using 25 of 26 letters, we are able to skip one letter at some point in each solution
     - So for solutions that contain a 'q' in the first word eg "quack", the next word must contain an 'x' or a 'j'
     - For solutions that contain an 'x', e.g. 'fldxt' in the first word, we have already skipped 'q' so the next word must contain 'j'

#### Breadth-wise solution
 - We solve for solutions that contain a 'q' in the first word and words that contain 'x' in the first word separately
   - solving separately reduces the problem set significantly which is much better for CPU and memory management
   - it also allows us to solve the solutions starting with an 'x' word faster because we are only looking for next possible words that contain the lowest unused letter rather than the next two unused letters as with 'q'

 - Given that the next lowest unused letter can be different depending on the starting word
   - we save time by filtering on all the next possible words for all the starting set
   - and then tackle the solution in one operation
   - this provides a better balance better iteratively solving for just one solution at a time whilst still allowing us to prune aggressively

 - For all the given possible solutions, any that do not contain the next lowest unused letter or a distinct combination of letters are discarded
   - this leaves us with the starting point to solve for the next word

 - For each iteration of the solution between 2-5 words, we reduce what we need to solve for by:
   - first of all we combine all the partial solutions into one combined word
   - and then we find the unique list of partial solutions
   - then before returning the result, we un-deduplicate and un-combine the words

Finally, before saving results,
we look up the anagrams that we removed at the beginning to end up with a complete set of solutions