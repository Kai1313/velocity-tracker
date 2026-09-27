package handler

import (
	"context"
	"net/http"
)

// The 5 entity handlers (User, Project, Ticket, Sprint, SprintEntry) all
// follow the same Create/Get/List/Update/Delete shape; only the request/
// response types and the service call itself differ per entity. These
// helpers factor out that shape, leaving each handler method as a one-line
// call adapting its own service signature to the shared flow.

func handleCreate[Req, Resp any](w http.ResponseWriter, r *http.Request, create func(context.Context, Req) (Resp, error)) {
	var req Req
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	resp, err := create(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func handleList[Resp any](w http.ResponseWriter, r *http.Request, list func(context.Context) ([]Resp, error)) {
	items, err := list(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func handleGet[Resp any](w http.ResponseWriter, r *http.Request, get func(context.Context, int64) (Resp, error)) {
	id, ok := pathID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	resp, err := get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleUpdate[Req, Resp any](w http.ResponseWriter, r *http.Request, update func(context.Context, int64, Req) (Resp, error)) {
	id, ok := pathID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req Req
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	resp, err := update(r.Context(), id, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleDelete(w http.ResponseWriter, r *http.Request, del func(context.Context, int64) error) {
	id, ok := pathID(r)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := del(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
