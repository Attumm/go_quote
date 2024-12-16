from dataclasses import dataclass
import unittest
from main import parse_text


@dataclass
class NoChangeTestCase:
    name: str
    expected: str


@dataclass
class TestCase:
    title: str
    test_input: str
    expected: str


test_cases = [
    TestCase("Basic 'im' correction", "im happy", "I'm happy."),
    TestCase("Uppercase 'IM' correction", "IM happy", "I'm happy."),
    TestCase("Capitalized 'Im' correction", "Im happy", "I'm happy."),
    TestCase("Basic 'i'm' correction", "i'm sad", "I'm sad."),
    TestCase("Uppercase 'I'M' correction", "I'M happy", "I'm happy."),
    TestCase("Sentence capitalization", "hello. how are you?", "Hello. How are you?"),
    TestCase("Multiple sentence correction", "hello there. im john.", "Hello there. I'm john."),
    TestCase("Standalone 'i' correction", "i am here", "I am here."),
    TestCase("'i' in middle of sentence", "you and i", "You and I."),
    TestCase("'i' with punctuation", "i. me. myself.", "I. Me. Myself."),
    TestCase("'i' at start of sentence", "i like python", "I like python."),
    TestCase("'i' in middle of sentence with space", "hello i am john", "Hello I am john."),
    TestCase("Multiple 'i' corrections", "im john. i like python. it's cool.", "I'm john. I like python. It's cool."),
    TestCase("Single 'i'", "i", "I"),
    TestCase("'i' with period", "i.", "I."),
    TestCase("'i' with question mark", "i?", "I?"),
    TestCase("'i' with exclamation mark", "i!", "I!"),
    TestCase("'i'm' with period", "i'm.", "I'm."),
    TestCase("'im' with period", "im.", "I'm."),
    TestCase("'i'm' with question mark", "i'm?", "I'm?"),
    TestCase("'im' with exclamation mark", "im!", "I'm!"),
    TestCase("Multiple 'i' in sentence", "i think i can i can", "I think I can I can."),
    TestCase("'i' with comma", "i, myself, and i", "I, myself, and I."),
    TestCase("'i' with semicolon", "i; however, i", "I; however, I."),
    TestCase("'i' in double quotes", 'he said "i am here"', 'He said "I am here".'),
    TestCase("'i' in single quotes", "she replied 'i know'", "She replied 'I know'."),
    TestCase("'i' with contraction 'd'", "i'd like to", "I'd like to."),
    TestCase("'i' with contraction 'll'", "i'll be there", "I'll be there."),
    TestCase("'i' with contraction 've'", "i've seen it", "I've seen it."),
    TestCase("'i' at end of sentence", "it was i.", "It was I."),
    TestCase("'i' at end of question", "who am i?", "Who am I?"),
    TestCase("Multiple sentences with 'i'", "i am here. you are there. i see you.", "I am here. You are there. I see you."),
    TestCase("'i' with numbers", "i have five apples", "I have five apples."),
    TestCase("'i' as part of a word", "there are 2 i's in this sentence", "There are 2 I's in this sentence."),
    TestCase("Multiple 'im' variations", "im gonna im going im gone!", "I'm gonna I'm going I'm gone!"),
    #TestCase("Mixed case 'i'", "i Am hErE. I aM tHeRe.", "I am here. I am there."),
    TestCase("'i' in words", "this is in italic?", "This is in italic?"),
    TestCase("Empty string", "", ""),
    TestCase("Whitespace only", "   ", "   "),
    TestCase("'i' with surrounding spaces", " i ", " I "),
    TestCase("Multiple spaces", "i  am   here", "I  am   here."),
    TestCase("No changes needed", "Hello. I'm john. I like python.", "Hello. I'm john. I like python."),

    TestCase("Quote", "The past is made out of facts... i guess the future is just hope.", "The past is made out of facts... I guess the future is just hope."),

    TestCase("Quote", "At the end the day because i believe so strongly in leadership, what i look for first, what i try to assess, is integrity.", "At the end the day because I believe so strongly in leadership, what I look for first, what I try to assess, is integrity."),

    TestCase("End sentence correctly", "In an age of rust, she comes up stainless steel", "In an age of rust, she comes up stainless steel."),
    TestCase("End sentence correclty", "In an age of rust, she comes up stainless steel!", "In an age of rust, she comes up stainless steel!"),
    TestCase("End sentence correclty", "In an age of rust, she comes up stainless steel?", "In an age of rust, she comes up stainless steel?"),
    TestCase("'i' with surrounding spaces", " i ", " I "),
]




no_change_test_cases = [
    NoChangeTestCase("Properly capitalized I", "I am here."),
    NoChangeTestCase("I'm already correct", "I'm going to the store."),
    NoChangeTestCase("Multiple sentences with correct I", "I like apples. I also like oranges."),
    NoChangeTestCase("I in quotes", 'She said "I am happy" and left.'),
    NoChangeTestCase("I with punctuation", "Am I? Yes, I am!"),
    NoChangeTestCase("I in contractions", "I'd like to go. I've been there before."),
    NoChangeTestCase("I within words", "The igloo was interesting."),
    #NoChangeTestCase("Mixed case sentence", "The iPad is on the table. I use it often."),
    #NoChangeTestCase("Correctly capitalized im", "The im in 'interim' should not change."),
    NoChangeTestCase("Correctly formatted whitespace", "  I  am  here  ."),
    NoChangeTestCase("Empty string", ""),
    NoChangeTestCase("Only whitespace", "   "),

    NoChangeTestCase("Quote", "How is it possible to know so much about a person and yet know nothing at all?"),
    NoChangeTestCase("Quote", "How beautiful would history have been if it could be written beforehand and then acted out like drama!"),
    NoChangeTestCase("Quote", "What you dislike in another take care to correct in yourself."),
    NoChangeTestCase("Quote", ""),
    NoChangeTestCase("Quote", ""),
]


class TestTextCorrection(unittest.TestCase):
    def test_corrections(self):
        for case in test_cases:
            with self.subTest(case.title):
                self.assertEqual(parse_text(case.test_input), case.expected)

    def test_no_changes(self):
        for case in no_change_test_cases:
            with self.subTest(case.name):
                self.assertEqual(parse_text(case.expected), case.expected)

if __name__ == '__main__':
    unittest.main()

