// Presentation helpers for CR 603.2d trigger doubling attribution.
// The server decides whether the attribution is public and sends the
// already safe name; the client only chooses the compact chip text.

export function doubledTriggerLabel(doubledBy?: string, doubledByName?: string): string | null {
  if (!doubledBy && !doubledByName) return null;
  return doubledByName ? `additional (${doubledByName})` : "additional trigger";
}
