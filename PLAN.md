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

A Go wordle engine has been implemented in pkg/wordlegameengine. A lightweight HTTP server has been added in main.go, listening on port 9111 with /api/evaluate POST endpoint accepting JSON ({&quot;solution&quot;: &quot;...&quot;, &quot;turns&quot;: [...], &quot;proposed_guess&quot;: &quot;...&quot;}) and returning dummy JSON matching spec: {&quot;game_status&quot;: &quot;...&quot;, &quot;turn_valid&quot;: bool, &quot;shortlist_reduction&quot;: {&quot;before&quot;: int, &quot;after&quot;: int, &quot;ratio&quot;: float}, &quot;feedback&quot;: &quot;...&quot;}. Dummy validation (length/lowercase). Tests pass. No caching yet.

## Current Iteration: Integrate Engine for Request Validation

### Acceptance Criteria

1. On startup, main.go calls wordlegameengine.LoadWordlists('./data'); log.Fatal if error.

2. Import &quot;./pkg/wordlegameengine&quot; and &quot;log&quot;.

3. In evaluateHandler:
   - Validate req.Solution: w, err := wordlegameengine.NewSolution(req.Solution); if err != nil { http.Error(err.Error(), 400) }; w.Validate() same.
   - If req.ProposedGuess != &quot;&quot;: same for NewWord(req.ProposedGuess).Validate()
   - For each turn in req.Turns: NewWord(turn.Guess).Validate()
   - Retain existing checks: solution/proposed/guess len==5 lowercase (but now superseded by New* ), feedback len==5.
   - If all valid, set TurnValid = true in response, return dummy as before.

4. Update tests to cover real word/non-word validation cases.

5. `go test ./...` passes.

6. Server responds 400 with descriptive error for invalid words.

### Tasks

- [ ] **go-coder**: Implement the validation integration as per AC above. Run tests.

- [ ] **qa-requirements**: Verify AC met, tests pass, validation uses engine correctly, dummy logic unchanged.

## Future Iterations

- Compute real game_status (won/lost/ongoing based on turns len and last feedback).

- Compute real feedback for proposed_guess against solution.

- Compute real shortlist reduction using Game logic.

- Implement caching for first turn.

