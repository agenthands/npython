# Workflow: Conductor Spec-Driven Development

## Development Cycle (TDD)
All development must follow this 7-step TDD cycle:

1. **Write the Test First**: Test file must exist and fail (Confirm RED state).
2. **Run the Test**: Confirm it fails as expected.
3. **Write Minimal Implementation**: Just enough code to pass the test.
4. **Run the Test**: Confirm it passes (Confirm GREEN state).
5. **Refactor**: Clean up the code while keeping tests green.
6. **Run All Tests**: Ensure no regressions and high coverage.
7. **Commit**: Use descriptive commit messages.

## Phase Completion Verification and Checkpointing Protocol
At the end of each Phase, a manual verification must be performed to ensure all requirements in the specification have been met and the code is stable.

- **Checklist**:
    - [ ] All tests for the phase are passing.
    - [ ] Code coverage is 100% for new logic.
    - [ ] Linter and type-checker pass.
    - [ ] Documentation is updated if necessary.
