# go-test-logger - AI Documentation

Progressive loading map for AI assistants working with go-test-logger package.

**Entry Point**: This file should be referenced from CLAUDE.md.

## Package Overview

**Purpose**: Test logging utilities for Ginkgo/Gomega BDD tests

**Key Features**:

- ExpectErrorLog - Capture and validate expected error patterns, hide from output
- ConfigureTestLogging - Suite-level logging with LOG_LEVEL environment variable
- WithCapturedLogger - Manual log capture for custom validation
- AssertNoErrorLogs - Negative assertions for successful operations
- Pattern matching - Validate log patterns with regex support
- JSON support - Full support for structured JSON log validation

## Always Load

- `.ai/llms.md` (this file)

## Load for Complex Tasks

- `.ai/memory.md` - Design decisions, gotchas, backward compatibility notes
- `.ai/context.md` - Current changes (if exists and is current)

## Common Standards (Portable Patterns)

**See** `.ai/common/common-llms.md` for the complete list of common standards.

Load these common standards when working on this package:

### Core Go Patterns

- `common/standards/go/constructors.md` - New* constructor functions
- `common/standards/go/error-wrapping.md` - Error wrapping with %w
- `common/standards/go/type-organization.md` - Interface and type placement

### Testing

- `common/standards/testing/bdd-testing.md` - Ginkgo/Gomega patterns
- `common/standards/testing/test-categories.md` - Test organization
- `common/standards/testing/gomega-matchers.md` - Gomega assertion patterns

### Documentation

- `common/standards/documentation/pattern-documentation.md` - Documentation structure
- `common/standards/documentation/code-references.md` - Code examples

## Project Standards (Package-Specific)

This package has minimal package-specific standards since it provides testing utilities.

Any package-specific patterns should go in `.ai/project-standards/`

## Loading Strategy

| Task Type | Load These Standards |
|-----------|---------------------|
| Adding new logging function | constructors.md, type-organization.md, bdd-testing.md |
| Writing tests | bdd-testing.md, test-categories.md, gomega-matchers.md |
| Documenting utilities | pattern-documentation.md, code-references.md |
| Ensuring compatibility | memory.md (for backward compatibility notes) |

## File Organization

```
go-test-logger/
├── CLAUDE.md                   # Entry point
├── .gitignore                  # Ignores context.md, memory.md, tasks/
└── .ai/
    ├── llms.md                 # This file (loading map)
    ├── README.md               # Documentation about .ai setup
    ├── context.md              # Current work (gitignored)
    ├── memory.md               # Stable knowledge (gitignored)
    ├── tasks/                  # Scratchpad (gitignored)
    ├── project-standards/      # Package-specific (if needed)
    └── common -> ~/code/ai-common  # Symlink to shared standards
```

## Key Principles

1. **Backward Compatibility**: Never break existing logging utilities or behavior
2. **Generic Design**: No project-specific logging patterns in this package
3. **Gomega Integration**: Works seamlessly with Gomega matchers (gbytes.Buffer, gbytes.Say)
4. **Clean Test Output**: Expected errors validated but hidden from output
5. **Environment Aware**: Respects LOG_LEVEL for debugging

## Related Documentation

- Common standard: `common/standards/testing/test-logging.md` - How to USE this package
- This is the implementation, that is the usage guide
