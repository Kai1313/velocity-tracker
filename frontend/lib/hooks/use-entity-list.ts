'use client';

import { useEffect, useState } from 'react';

// The simple admin list pages (users, projects, sprints, tickets) all fetch
// a collection once on mount, then keep it in sync locally after a create/
// update (upsert) or delete (remove) instead of refetching. This factors
// that shape out; each page still owns its own fetch call and delete API
// call, since those differ per entity.
export function useEntityList<T extends { id: number }>(fetcher: () => Promise<T[]>, errorMessage: string) {
  const [items, setItems] = useState<T[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetcher()
      .then(setItems)
      .catch((err) => setError(err instanceof Error ? err.message : errorMessage));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  function upsert(item: T) {
    setItems((prev) => {
      if (!prev) return [item];
      const exists = prev.some((i) => i.id === item.id);
      return exists ? prev.map((i) => (i.id === item.id ? item : i)) : [...prev, item];
    });
  }

  function remove(id: number) {
    setItems((prev) => prev?.filter((i) => i.id !== id) ?? null);
  }

  return { items, error, upsert, remove };
}
