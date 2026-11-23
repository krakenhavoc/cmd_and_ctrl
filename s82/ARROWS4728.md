# Mockups for the combat arrows redesign (#4728)

Ian filed #4728 from the app: the attack/block signaling ("combat arrows") during combat looks ugly. #4813 already fixed the rendering bug where the arrows drew as solid black shapes. What's left is the visual design. Ian wants **2 or 3 mockups to choose from. Do not build the real feature.**

Read DESIGN.md (the UI design brief), AGENTS.md §5 (labels, design skills; `.claude/skills/web-design-guidelines`), #4728 with its screenshot and comments, #4327 ("Intent to attack", a related player report), and the client code that draws the combat arrows today (grep `client/src` for the combat arrow or attack line component).

1. On a branch `design/4728-combat-arrows` off origin/develop, add `docs/design/4728-combat-arrows/`. In it, write 2–3 self-contained static HTML mockups. Each shows a 4-player table mid-combat:
   - one attacker each at two different opponents;
   - one creature blocked by two blockers;
   - one unblocked creature.
   Show it in light and dark themes, using DESIGN.md's tokens and the client's real colours and type.
   Each mockup is a distinct direction, for example:
   - curved, coloured lines from attacker to defending player or planeswalker, with block lines in a second style;
   - no lines at all: badges or tags on the cards ("→ Luke", "blocking Grizzly Bears") plus a hover highlight;
   - a hybrid.
   Cover: who is being attacked, which blockers block which attacker, multiple blockers, colour-blind safety, phone width, and how it reads in the fixed table layout.
2. Render each mockup to PNG with Playwright (the repo has it in tests-e2e). Commit the PNGs beside the HTML.
3. Push the branch. Do not open a PR. Comment once on #4728 with:
   - each option's PNG, linked from the branch through raw.githubusercontent.com URLs;
   - a short paragraph per option: what it does well, and its costs to build, including what client code it touches;
   - your recommendation;
   - and finally: "Pick one (or a mix) and an agent will build it."
4. Change no app code.

Final message: the branch, the comment URL and a one-line summary of each option.

Nobody will answer you mid-run.
