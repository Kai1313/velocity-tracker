'use client';

import { useState } from 'react';

// Every admin form dialog (create/edit) repeats the same open/pending/error
// state and the same submit-then-close-or-show-error flow. Field state and
// the reset-on-open effect stay in each dialog, since those are specific to
// what the entity's form actually holds.
export function useFormDialogState() {
  const [open, setOpen] = useState(false);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function submit<T>(save: () => Promise<T>, onSaved: (value: T) => void) {
    setPending(true);
    setError(null);
    try {
      const saved = await save();
      onSaved(saved);
      setOpen(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Save failed');
    } finally {
      setPending(false);
    }
  }

  return { open, setOpen, pending, error, setError, submit };
}
