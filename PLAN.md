# Project Plan

## Overall Goal

We need a Go service listening on port 9111. It will be receiving POST requests from a local Python programme, which is a machine learning model. The ML model is using reinforcement learning to learn to play Wordle.

This Go application has knowledge of the Wordle wordlists for valid guesses, and valid solutions, in the `./data/` folder.

The API will be very simple, with no extraneous features or fields. There will not be versioning.

### Input

When a request comes in, the details in the request should be the correct solution, and between 0-5 turns. A turn will have the word played, and the feedback. Feedback looks like 'GY--Y', where G = green tile in a Wordle game, Y = yellow tile, '-' = grey

### Output

When sent data about the state of a game to date, and a proposed new move, the Go service responds with a JSON body containing the following information:
- Whether the game is won, lost or ongoing after the latest turn. You get 6 turns.
- Whether the turn is valid or invalid. It should have been a 5-letter word in lower case on the allowed guess list
- The 'solution shortlist reduction'. This is the number of possible solutions before the latest turn, the number of possible solutions after the latest turn, and the reduction on a scale of 0-1.
- The 'feedback'.

### Caching

There will be a large, in-memory cache. It will be used for identifying the solution shortlist after the first turn in a game only. It will not be possible to cache solution shortlists for subsequent terms. The cache keys are the first turn the player took, and the feedback string. The value is the entire remaining shortlist, which will often contain hundreds or thousands of items. No more than 10GB should be used for the cache, roughly

## Progress so far

A Go wordle engine has been implemented, either completely or at least partially. The application layer hasn't been done. There is no caching.


## Remaining Development Plan

# TODO - write development plan, in order, for features to add