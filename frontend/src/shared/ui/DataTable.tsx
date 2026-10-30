import type { ReactNode } from "react";

export type DataTableColumn<Row> = {
  key: string;
  header: string;
  sortKey?: string;
  render: (row: Row) => ReactNode;
};

type DataTableProps<Row> = {
  caption: string;
  columns: DataTableColumn<Row>[];
  rows: Row[];
  rowKey: (row: Row) => string;
  loading: boolean;
  error?: string;
  empty: boolean;
  hasFilters: boolean;
  onRetry: () => void;
  onResetFilters: () => void;
  emptyAction: ReactNode;
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  sortBy: string;
  sortOrder: "asc" | "desc";
  onSort: (sortBy: string, sortOrder: "asc" | "desc") => void;
};

export function DataTable<Row>({
  caption, columns, rows, rowKey, loading, error, empty, hasFilters, onRetry, onResetFilters,
  emptyAction, page, pageSize, total, onPageChange, sortBy, sortOrder, onSort,
}: DataTableProps<Row>) {
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  if (loading) {
    return (
      <div className="table-state table-loading" role="status" aria-label="Cargando registros">
        {Array.from({ length: 5 }, (_, row) => (
          <div className="table-skeleton-row" key={row}>
            {columns.map((column) => <span className="skeleton-block" key={column.key} />)}
          </div>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="table-state table-error" role="alert">
        <p>{error}</p>
        <button className="secondary-button" type="button" onClick={onRetry}>Reintentar</button>
      </div>
    );
  }

  if (empty) {
    return (
      <div className="table-state table-empty">
        <h2>{hasFilters ? "Sin resultados" : "Aún no hay usuarios"}</h2>
        <p>{hasFilters ? "No encontramos usuarios con esos criterios. Ajusta la búsqueda o los filtros." : "Registra el primer usuario para empezar a administrar las cuentas."}</p>
        {hasFilters ? <button className="secondary-button" type="button" onClick={onResetFilters}>Limpiar búsqueda y filtros</button> : emptyAction}
      </div>
    );
  }

  return (
    <>
      <div className="users-table-container">
        <table className="users-table">
          <caption className="visually-hidden">{caption}</caption>
          <thead>
            <tr>
              {columns.map((column) => {
                const active = column.sortKey === sortBy;
                const ariaSort = active ? (sortOrder === "asc" ? "ascending" : "descending") : "none";
                return (
                  <th key={column.key} scope="col" aria-sort={column.sortKey ? ariaSort : undefined}>
                    {column.sortKey ? (
                      <button className="table-sort-button" type="button" onClick={() => onSort(column.sortKey!, active && sortOrder === "asc" ? "desc" : "asc")}>
                        {column.header}<span aria-hidden="true">{active ? (sortOrder === "asc" ? " ↑" : " ↓") : " ↕"}</span>
                      </button>
                    ) : column.header}
                  </th>
                );
              })}
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => <tr key={rowKey(row)}>{columns.map((column) => <td key={column.key}>{column.render(row)}</td>)}</tr>)}
          </tbody>
        </table>
      </div>
      <nav className="table-pagination" aria-label="Paginación de usuarios">
        <span>{total === 0 ? 0 : (page - 1) * pageSize + 1}–{Math.min(page * pageSize, total)} de {total}</span>
        <div>
          <button className="secondary-button" type="button" onClick={() => onPageChange(page - 1)} disabled={page <= 1}>Anterior</button>
          <span>Página {page} de {totalPages}</span>
          <button className="secondary-button" type="button" onClick={() => onPageChange(page + 1)} disabled={page >= totalPages}>Siguiente</button>
        </div>
      </nav>
    </>
  );
}
