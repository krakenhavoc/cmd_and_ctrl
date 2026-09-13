// botChat turns the raw chat stream into the table feed of bot
// disclosures. S31 sub-PR 8 / ADR 0033 §8.
//
// Two kinds of line arrive from a bot seat and they are NOT the same
// thing:
//
//   bot_improvisation — the bot applied an effect by hand because the
//     rules engine can't run that card. This is mandatory disclosure.
//     It is never filtered, never collapsed, and no setting hides it:
//     an unannounced improvisation is a bot cheating, and a client
//     that suppresses the announcement is the thing doing the
//     cheating. `visibleBotLines` has no code path that drops one.
//
//   bot_reasoning — the bot narrating why it took an ordinary move.
//     Debug output. Hidden unless the viewer turned on
//     settings.gameplay.showBotReasoning.
//
// The reasoning attached to an improvisation follows the same rule as
// a reasoning line: shown only behind the setting. The announcement
// itself is complete without it.
//
// Kept as plain functions over plain data so the rules above are
// unit-testable without mounting a component.

import { CHAT_BOT_IMPROVISATION, CHAT_BOT_REASONING } from "./protocol";
import type { ChatMessage } from "./ws";

// BOT_FEED_LIMIT caps how many lines the on-table feed shows at once.
// The feed sits in the board's attention strip, which is shared with
// the targeting banner and the toasts, so it has to stay small. The
// full history is in the chat store and the replay log.
export const BOT_FEED_LIMIT = 4;

export interface BotLine {
  id: string;
  authorID: string;
  authorName: string;
  text: string;
  // reason is empty unless the viewer asked to see reasoning.
  reason: string;
  improvised: boolean;
  at: Date;
}

// isBotLine reports whether a chat message came from a bot seat
// rather than from a person.
export function isBotLine(msg: ChatMessage): boolean {
  return msg.kind === CHAT_BOT_IMPROVISATION || msg.kind === CHAT_BOT_REASONING;
}

// isImprovisation reports whether a message is a mandatory
// improvisation disclosure.
export function isImprovisation(msg: ChatMessage): boolean {
  return msg.kind === CHAT_BOT_IMPROVISATION;
}

// visibleBotLines projects the chat log into the feed, newest last,
// capped at `limit`.
//
// `showReasoning` adds the bot_reasoning lines and attaches the
// `reason` field to the improvisations. It never removes an
// improvisation.
export function visibleBotLines(
  chat: readonly ChatMessage[],
  showReasoning: boolean,
  limit: number = BOT_FEED_LIMIT,
): BotLine[] {
  const out: BotLine[] = [];
  for (const msg of chat) {
    if (!isBotLine(msg)) continue;
    const improvised = isImprovisation(msg);
    if (!improvised && !showReasoning) continue;
    out.push({
      id: msg.id,
      authorID: msg.authorID,
      authorName: msg.authorName,
      text: msg.text,
      reason: showReasoning ? msg.reason : "",
      improvised,
      at: msg.at,
    });
  }
  return limit > 0 && out.length > limit ? out.slice(-limit) : out;
}

// dismissable reports whether the viewer may clear a line from the
// feed. Reasoning is chatter and can go; an improvisation cannot be
// dismissed before it has been seen, so the feed ages it out on the
// cap rather than offering a close button. (The undo, not a
// dismissal, is the response to an improvisation you disagree with.)
export function dismissable(line: BotLine): boolean {
  return !line.improvised;
}
