package methodologies

// TDDMethodology contains Test-Driven Development guidance.
const TDDMethodology = `# Test-Driven Development

## The Iron Law

NO PRODUCTION CODE WITHOUT A FAILING TEST FIRST

Write code before the test? Delete it. Start over. No exceptions.

## Red-Green-Refactor Cycle

### RED - Write Failing Test

Write one minimal test showing what should happen.

Requirements:
- One behavior per test
- Clear descriptive name
- Real code (no mocks unless unavoidable)

### Verify RED - Watch It Fail

MANDATORY. Never skip.

Run the test. Confirm:
- Test fails (not errors)
- Failure message is expected
- Fails because feature missing (not typos)

Test passes immediately? You're testing existing behavior. Fix test.

### GREEN - Minimal Code

Write simplest code to pass the test.

Don't add features, refactor other code, or "improve" beyond the test.

### Verify GREEN - Watch It Pass

MANDATORY.

Run the test. Confirm:
- Test passes
- Other tests still pass
- Output clean (no errors, warnings)

Test fails? Fix code, not test.

### REFACTOR - Clean Up

After green only:
- Remove duplication
- Improve names
- Extract helpers

Keep tests green. Don't add behavior.

## Common Rationalizations - All Mean DELETE AND START OVER

| Excuse | Reality |
|--------|---------|
| "Too simple to test" | Simple code breaks. Test takes 30 seconds. |
| "I'll test after" | Tests passing immediately prove nothing. |
| "Already manually tested" | Ad-hoc ≠ systematic. No record, can't re-run. |
| "Keep as reference" | You'll adapt it. That's testing after. Delete means delete. |
| "Need to explore first" | Fine. Throw away exploration, start with TDD. |

## Red Flags - STOP and Start Over

- Code written before test
- Test passes immediately
- Can't explain why test failed
- Rationalizing "just this once"
- "Tests after achieve the same purpose"

## Verification Checklist

Before claiming task complete:
- Every new function/method has a test
- Watched each test fail before implementing
- Wrote minimal code to pass each test
- All tests pass
- Output clean (no errors, warnings)

Can't check all boxes? You skipped TDD. Start over.`
