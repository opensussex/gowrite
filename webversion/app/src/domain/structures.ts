import type { Chapter } from "./models";

function chapter(title: string, notes: string, content: string): Omit<Chapter, "id"> {
  return {
    title,
    notes,
    content,
    target: 0
  };
}

export const STRUCTURE_TEMPLATES: Record<string, Omit<Chapter, "id">[]> = {
  "3act": [
    chapter(
      "Act 1: The Setup",
      "Introduce characters and the ordinary world.",
      ">> GUIDANCE: Introduce your protagonist and the status quo."
    ),
    chapter(
      "Inciting Incident",
      "Something disrupts normal life.",
      ">> GUIDANCE: Trigger the event that cannot be ignored."
    ),
    chapter(
      "Plot Point 1",
      "Hero commits to the journey.",
      ">> GUIDANCE: Move from comfort to conflict."
    ),
    chapter(
      "Act 2: Confrontation",
      "Rising action, tests, allies, enemies.",
      ">> GUIDANCE: Increase pressure and complexity."
    ),
    chapter(
      "Midpoint",
      "A major shift: false victory or defeat.",
      ">> GUIDANCE: Raise stakes and lock in consequences."
    ),
    chapter(
      "Plot Point 2",
      "All seems lost; force transformation.",
      ">> GUIDANCE: Push protagonist to change."
    ),
    chapter(
      "Act 3: Resolution",
      "Final confrontation and climax.",
      ">> GUIDANCE: Resolve central conflict with earned choices."
    ),
    chapter("The End", "Show aftermath and new normal.", ">> GUIDANCE: Show who changed and why it matters.")
  ],
  hero: [
    chapter("Ordinary World", "Status quo.", ">> GUIDANCE: Show life before the journey."),
    chapter("Call to Adventure", "Disruption appears.", ">> GUIDANCE: Present challenge/opportunity."),
    chapter("Refusal of the Call", "Fear and hesitation.", ">> GUIDANCE: Why resist the journey?"),
    chapter("Meeting the Mentor", "Tools and confidence.", ">> GUIDANCE: Prepare the hero."),
    chapter("Crossing the Threshold", "Commitment to new world.", ">> GUIDANCE: No turning back."),
    chapter("Tests, Allies, Enemies", "Learn new rules.", ">> GUIDANCE: Build conflict ecosystem."),
    chapter("The Ordeal", "Major internal/external trial.", ">> GUIDANCE: Confront deepest fear."),
    chapter("Return with Elixir", "Transformation complete.", ">> GUIDANCE: Bring change home.")
  ],
  cat: [
    chapter("Opening Image", "Snapshot before change.", ">> GUIDANCE: Establish tone quickly."),
    chapter("Theme Stated", "What story is truly about.", ">> GUIDANCE: State theme indirectly."),
    chapter("Setup", "Show flaw and stakes.", ">> GUIDANCE: Plant all key story threads."),
    chapter("Catalyst", "Life changes forever.", ">> GUIDANCE: Force story motion."),
    chapter("Debate", "Should I do this?", ">> GUIDANCE: Tension between fear and action."),
    chapter("Break into Two", "Enter new world.", ">> GUIDANCE: Active choice starts Act 2."),
    chapter("Midpoint", "False win/loss.", ">> GUIDANCE: Reframe objective and stakes."),
    chapter("Finale", "Execute final plan.", ">> GUIDANCE: Prove transformation.")
  ],
  fichtean: [
    chapter("Inciting Incident", "Start with disruption.", ">> GUIDANCE: Skip long setup."),
    chapter("Crisis 1", "First obstacle.", ">> GUIDANCE: Problem escalates."),
    chapter("Crisis 2", "Higher stakes.", ">> GUIDANCE: Pressure increases."),
    chapter("Crisis 3", "Near breaking point.", ">> GUIDANCE: Strip options away."),
    chapter("Climax", "Maximum tension.", ">> GUIDANCE: Resolve core conflict."),
    chapter("Resolution", "Aftermath.", ">> GUIDANCE: Establish new normal.")
  ],
  horror: [
    chapter("The Dreadful Normal", "Normal life with unease.", ">> GUIDANCE: Seed dread early."),
    chapter("The Omen", "Warning sign appears.", ">> GUIDANCE: Hint at threat."),
    chapter("The Onset", "Threat reveals itself.", ">> GUIDANCE: First undeniable horror."),
    chapter("The Discovery", "Characters understand danger.", ">> GUIDANCE: Clarify monster/rules."),
    chapter("The Pursuit", "Escalating attacks.", ">> GUIDANCE: Constrict options and safety."),
    chapter("The Confrontation", "Final stand.", ">> GUIDANCE: Survive or fail."),
    chapter("The Aftermath", "Cost of survival.", ">> GUIDANCE: End with consequence.")
  ]
};

export function getStructureNames(): string[] {
  return Object.keys(STRUCTURE_TEMPLATES);
}
