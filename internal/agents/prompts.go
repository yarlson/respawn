package agents

import (
	"fmt"
	"strings"
)

const promptBlockSeparator = "\n\n---\n\n"

const agentsGeneratorRole = `You are an AGENTS.md generator. Create comprehensive agent guidelines from PRDs using progressive disclosure.

Your Task
Create the following files in the repository:
1. AGENTS.md - Minimal root file with project overview and links to docs/
2. docs/*.md - Domain-specific documentation files
3. CLAUDE.md - A symlink pointing to AGENTS.md
   - Use platform-appropriate commands: ln -s on Unix/macOS, mklink on Windows
   - If symlink creation fails, copy the file instead as a fallback

You MUST write these files directly using your file writing tools. Do NOT output file contents as text.

Feedback Loop Selection (CRITICAL)
Analyze the PRD to determine what type of work this project involves, then prescribe the appropriate development methodology and feedback loops:

**Backend/API/Library code:**
- MUST use Test-Driven Development (TDD)
- Red-Green-Refactor cycle: write failing test first, implement to pass, refactor
- Tests are the feedback loop - they give the agent "eyes" to know if code works
- Document TDD requirement prominently in AGENTS.md

**Frontend/UI code:**
- Use browser/UI validation as feedback loop
- Agent should verify UI renders correctly, buttons exist, layouts work
- Recommend visual regression testing or screenshot comparison
- Document UI verification patterns

**CLI tools:**
- Use output verification as feedback loop
- Test expected stdout/stderr output
- Verify exit codes
- Document CLI testing patterns

**Mixed projects:**
- Apply appropriate methodology to each component
- Backend parts get TDD, frontend gets UI validation
- Document both approaches

Progressive Disclosure Principles
1. **Root AGENTS.md should be minimal (<=300 lines)**
   - One-sentence project description
   - Development methodology (TDD, UI testing, etc.) - THIS IS CRITICAL
   - Package manager (if non-standard)
   - Build/test commands
   - Links to detailed documentation files in docs/

2. **Group related guidelines into separate docs/ files**
   - Go conventions -> docs/GO_CONVENTIONS.md
   - TypeScript conventions -> docs/TYPESCRIPT.md
   - Testing patterns -> docs/TESTING.md (ALWAYS create this)
   - API design -> docs/API_CONVENTIONS.md
   - Architecture decisions -> docs/ARCHITECTURE.md
   - Safety/security guardrails -> docs/SAFETY.md
   - Create only files relevant to the project

3. **Use markdown links for progressive disclosure**
   - In AGENTS.md: "For testing patterns, see [docs/TESTING.md](docs/TESTING.md)"
   - Each document stays focused on one domain

4. **Document capabilities, not file paths**
   - File paths change; capabilities are stable

5. **Write natural language**
   - Include concrete examples
   - Keep language conversational

Required Actions
1. Analyze the PRD to determine project type (backend/frontend/CLI/mixed)
2. Create docs/ directory: mkdir -p docs
3. Write AGENTS.md with appropriate methodology prominently stated
4. Write docs/TESTING.md with the correct feedback loop approach
5. Write other relevant docs/*.md files
6. Create CLAUDE.md symlink: ln -sf AGENTS.md CLAUDE.md

BEGIN
Read the PRD below, determine the appropriate development methodology, and create all required files now.`

func buildAgentsPrompt(prdContent string) string {
	return joinPromptBlocks(agentsGeneratorRole, agentsUserPrompt(prdContent))
}

func agentsUserPrompt(prdContent string) string {
	return fmt.Sprintf("PRD Content:\n%s\n\nAnalyze this PRD and create appropriate guidelines:\n\n1. FIRST: Determine project type and select methodology:\n   - Backend/API/Library -> TDD (Test-Driven Development)\n   - Frontend/UI -> Browser validation, visual testing\n   - CLI tools -> Output/exit code verification\n   - Mixed -> Apply appropriate method to each component\n\n2. AGENTS.md (in repository root)\n   - One-sentence project description\n   - **Development Methodology section** - state TDD or other approach clearly\n   - Core stack/technologies\n   - Primary commands (build, test, run)\n   - Links to docs/ files\n   - Keep it minimal (<=300 lines)\n\n3. docs/TESTING.md (REQUIRED)\n   - Describe the feedback loop approach for this project type\n   - For TDD: explain Red-Green-Refactor cycle\n   - For UI: explain browser/visual validation\n   - For CLI: explain output verification\n   - Include concrete examples\n\n4. Other docs/*.md files (create only what's relevant)\n   - docs/GO_CONVENTIONS.md (if Go project)\n   - docs/TYPESCRIPT.md (if TypeScript project)\n   - docs/ARCHITECTURE.md (system design)\n   - docs/SAFETY.md (security guardrails)\n   - docs/API_CONVENTIONS.md (if has APIs)\n\n5. CLAUDE.md symlink\n   - On macOS/Linux: Run: ln -sf AGENTS.md CLAUDE.md\n   - On Windows: Run: mklink CLAUDE.md AGENTS.md (or copy if mklink unavailable)\n   - The symlink must point to AGENTS.md so tools can find project guidelines\n\nWrite all files now using your tools. Do not output file contents as text.", prdContent)
}

func joinPromptBlocks(blocks ...string) string {
	trimmed := make([]string, 0, len(blocks))
	for _, block := range blocks {
		value := strings.TrimSpace(block)
		if value == "" {
			continue
		}
		trimmed = append(trimmed, value)
	}
	return strings.Join(trimmed, promptBlockSeparator)
}
