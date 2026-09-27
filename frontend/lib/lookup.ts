// Nearly every admin page re-derives a display label for an id by finding it
// in a related list and falling back to a placeholder when it's missing.
// This factors out that one-line pattern instead of reimplementing it per page.
export function lookupLabel<T extends { id: number }>(
  items: T[],
  id: number,
  label: (item: T) => string,
  fallback: string,
): string {
  const item = items.find((i) => i.id === id);
  return item ? label(item) : fallback;
}
