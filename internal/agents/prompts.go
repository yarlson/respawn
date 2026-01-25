package agents

import "fmt"

const agentsPromptTemplate = `You are an AGENTS.md generator. Create concise, general guidelines from PRDs.

Primary Constraints
- Keep content helpful but general.
- Do NOT include code snippets, directory trees, or file listings.
- Command examples are OK (tests, builds, linters), but keep them minimal.
- Avoid prescribing specific file paths inside the docs.

Your Task
Create the following files in the repository:
1. AGENTS.md - Minimal root guidelines and links to deeper docs
2. docs/TESTING.md - REQUIRED, describes feedback loops
3. Other focused docs only if relevant (architecture, API conventions, safety, language conventions)
4. CLAUDE.md pointing to AGENTS.md (symlink or copy)

Feedback Loop Selection (CRITICAL)
Analyze the PRD to determine what type of work this project involves, then prescribe the appropriate development methodology and feedback loops:

**Backend/API/Library code:**
- MUST use Test-Driven Development (TDD)
- Red-Green-Refactor cycle: write failing test first, implement to pass, refactor
- Tests are the feedback loop

**Frontend/UI code:**
- Use browser/UI validation as feedback loop
- Validate layout, interactions, and visual correctness
- Recommend visual regression testing concepts (without tool specifics)

**CLI tools:**
- Use output verification as feedback loop
- Verify exit codes and expected outputs (conceptually, no commands)

**Mixed projects:**
- Apply appropriate methodology to each component

Progressive Disclosure Principles
1. **Root AGENTS.md should be minimal**
   - One-sentence project description
   - Development methodology (TDD, UI validation, CLI verification)
   - How to decide what to verify
   - Links to deeper docs

2. **Keep docs focused**
   - Testing patterns -> docs/TESTING.md
   - Architecture decisions -> docs/ARCHITECTURE.md
   - Safety/security guardrails -> docs/SAFETY.md
   - API conventions -> docs/API.md
   - Language conventions -> docs/GO.md or docs/TYPESCRIPT.md
   - Create only files relevant to the project

3. **Write natural language**
   - Short, actionable paragraphs and bullets
   - Avoid code blocks; simple command examples are OK

BEGIN
Read the PRD below, determine the appropriate development methodology, and create all required files now.

---

PRD Content:
%s

Analyze this PRD and create appropriate guidelines.

Content requirements:
- Keep docs general and helpful, avoid implementation details.
- Do NOT include code snippets, directory trees, or file listings.
- Command examples are OK (tests, builds, linters), but keep them minimal.
- Avoid specific file paths in the doc content.

Required outputs:
1. AGENTS.md with a concise project overview and methodology.
2. docs/TESTING.md describing the feedback loop approach for this project type.
3. Additional focused docs only if truly relevant (architecture, safety, API conventions, language conventions).
4. CLAUDE.md pointing to AGENTS.md.

Write all files now using your tools. Do not output file contents as text.`

func buildAgentsPrompt(prdContent string) string {
	return fmt.Sprintf(agentsPromptTemplate, prdContent)
}
