from itertools import product
import numpy.typing as npt
import numpy as np
import time


def read_word_file(file_path: str) -> (npt.NDArray[str], npt.NDArray[np.int32], npt.NDArray[np.int32]):
    """read word list and prepare for solving"""
    with open(file_path, 'r') as f:
        # remove /n
        all_words_raw = map(lambda x: x[: -1], f)
        # five-letter words
        all_words_raw = filter(lambda x: len(x) == WORD_LEN, all_words_raw)
        # five unique letters
        all_words_raw = list(filter(lambda x: len(set(x)) == WORD_LEN, all_words_raw))
        # encode words
        words_encoded = list(map(encode_word, all_words_raw))
        # convert to numpy array
        all_words_raw = np.array(all_words_raw)
        all_words = np.array(words_encoded, dtype=np.int32)
        # remove anagrams for solving
        words = np.unique(all_words).reshape((-1, 1))
    return all_words_raw, all_words, words


def write_results(results: npt.NDArray[np.int32], file_path: str):
    """unpack anagrams and save to file with comma separated values"""
    output = []
    for row in results:
        solutions = [ALL_WORDS_RAW[ALL_WORDS == word] for word in row]
        solutions = list(product(*solutions))
        for word_set in solutions:
            output.append(','.join(word_set))

    with open(file_path, 'w') as f:
        f.write('\n'.join(output))


def encode_word(word: str) -> int:
    """coverts a word to a 26-bit integer representation"""
    char_values = [1 << (ord(char) - ord('a')) for char in word]
    return sum(char_values)


def chars_from_words(a: npt.NDArray[np.int32]) -> npt.NDArray[np.int32]:
    """
    returns the int representation of each letter in a word as a list
    :param a: 1d array of words or combined words (all the same character length)
    :return: 2d array with columns for each character, characters ordered by lowest to the highest frequency
    """
    chars_index = np.where(REVERSE_FREQ_ALPHABET & a != 0)[1].reshape((a.shape[0], -1))
    return REVERSE_FREQ_ALPHABET[chars_index]


def reverse_alphabet(word_list: npt.NDArray[np.int32]) -> npt.NDArray[np.int32]:
    """returns alphabet in reverse order by frequency of letters in word list"""
    alphabet = np.array([1 << i for i in range(26)], dtype=np.int32)
    freq = np.sum(alphabet & word_list == alphabet, axis=0)
    return alphabet[np.argsort(freq)]


def decode_word(word: int) -> str:
    """transforms integer representation of word to string"""
    chars = REVERSE_FREQ_ALPHABET[word & REVERSE_FREQ_ALPHABET != 0]
    indices = np.array(np.log2(chars) + ord('a'), dtype=np.int32)
    letters = map(chr, indices)
    return ''.join(letters)


def find_next_chars_to_explore(a: npt.NDArray[np.int32], n_chars: int) -> npt.NDArray[np.int32]:
    """
    :param a: 1d array with valid partial solutions
    :param n_chars: the number of characters to explore in for the next solution where the char is the
    next un-utilised character by reverse order frequency in the overall word-set
    :return: 2d array with next characters, by frequency, to explore for each partial solution
    """
    # find all characters in each combined partial solution
    chars = chars_from_words(a)
    # find reverse alphabet index of the smallest char in each word (col0 is the smallest in each word by freq)
    min_char_index = np.where(chars[:, 0].reshape(-1, 1) == REVERSE_FREQ_ALPHABET)[1]

    # find the index of each character not in the word
    not_chars_index = np.where(REVERSE_FREQ_ALPHABET & a == 0)[1]
    # find the characters not in each word
    not_chars = REVERSE_FREQ_ALPHABET[not_chars_index].reshape((a.shape[0], -1))

    # rank characters in order they come after the minimum character (by freq) in word
    index = np.cumsum(not_chars_index.reshape((a.shape[0], -1)) > min_char_index.reshape((-1, 1)), axis=1)
    # find the two smallest letters of the letters in alphabet after smallest letter in word
    next_chars = not_chars[np.isin(index, np.arange(1, n_chars + 1))].reshape((-1, n_chars))
    return next_chars


def solve_next_word(a: npt.NDArray[np.int32], n_chars: int) -> npt.NDArray[np.int32]:
    """
    :param n_chars: the number of characters to explore in for the next solution where the char is the
    next un-utilised character by reverse order frequency in the overall word-set
    :param a: 2d array with valid partial solutions
    :return: 2d array with valid partial solutions with the next valid word
    """
    # combining words into a single integer
    a_combined = np.bitwise_xor.reduce(a, axis=1)
    # find unique partial solutions to solve for
    a_combined_unique, inverse = np.unique(a_combined, return_inverse=True)
    a_combined_unique = a_combined_unique.reshape((-1, 1))

    # find the next words to solve for
    next_chars = find_next_chars_to_explore(a_combined_unique, n_chars)
    next_words = WORDS[np.isin(WORD_MIN_CHARS, next_chars)]
    next_words = next_words[~np.isin(next_words, a[0])].reshape(1, -1)

    # solve
    bool_array = (
        # check words contain mutually exclusive letters - creates 2d boolean array
        (a_combined_unique & next_words == 0) &
        # trim solutions: check that word contains at least one of the next letters
        (np.bitwise_or.reduce(next_chars, axis=1).reshape((-1, 1)) & next_words > 0)
        )[inverse]

    # join next words to original partial solution
    results = np.concatenate((
        # each original partial solution is repeated for each next word that fits it
        a[np.repeat(np.arange(a.shape[0]), np.sum(bool_array, axis=1), axis=0)],
        # next words are flattened
        np.tile(next_words, a.shape[0])[0][bool_array.ravel()].reshape((-1, 1))
        ), axis=1)
    return results


def solve(start_letter: int, n_next_chars: int):
    """
    :param start_letter: by frequency of appearance in the overall word-set, the smallest letter in the starting
    words
    :param n_next_chars: the number of characters to explore for in the next solution where the char is the
    next un-utilised character by reverse order frequency in the overall word-set
    :return: 2d array of unique solutions
    """
    # first words in solution
    results = WORDS[np.isin(WORD_MIN_CHARS, REVERSE_FREQ_ALPHABET[start_letter: start_letter + 1])]
    # solve for words 2-5
    for _ in range(2, N_WORDS + 1):
        results = solve_next_word(results, n_chars=n_next_chars)
        if n_next_chars > 1:
            results = np.unique(np.sort(results, axis=1), axis=0)
    return results


if __name__ == '__main__':

    start = time.time()

    WORD_LEN = 5
    N_WORDS = 26 // WORD_LEN

    ALL_WORDS_RAW, ALL_WORDS, WORDS = read_word_file('words_alpha.txt')
    # alphabet in order of low to high frequency each letter appears in overall word set
    REVERSE_FREQ_ALPHABET = reverse_alphabet(WORDS)
    # the lowest character in each word by frequency it appears in overall word set
    WORD_MIN_CHARS = chars_from_words(WORDS)[:, 0]

    # first words contain - q
    sol1 = solve(start_letter=0, n_next_chars=2)

    # first words contain - x
    sol2 = solve(start_letter=1, n_next_chars=1)

    sol = np.concatenate((sol1, sol2), axis=0)
    sol = np.unique(np.sort(sol, axis=1), axis=0)

    write_results(sol, file_path='results.csv')
    print('total time: ', time.time() - start)
