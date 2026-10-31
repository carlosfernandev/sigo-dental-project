#!/usr/bin/env bash

set -Eeuo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
FRONTEND_DIR="${PROJECT_ROOT}/frontend"
FRONTEND_PORT=5173
API_HEALTH_URL="http://127.0.0.1:8080/health"
VITE_LOG="$(mktemp -t sigo-vite-ngrok.XXXXXX.log)"
VITE_PID=""

cleanup() {
  if [[ -n "${VITE_PID}" ]] && kill -0 "${VITE_PID}" 2>/dev/null; then
    kill "${VITE_PID}" 2>/dev/null || true
    wait "${VITE_PID}" 2>/dev/null || true
  fi
  rm -f -- "${VITE_LOG}"
}

trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

if ! command -v ngrok >/dev/null 2>&1; then
  echo "Error: ngrok no está instalado. Instálalo desde https://ngrok.com/download" >&2
  exit 1
fi

if ! command -v npm >/dev/null 2>&1; then
  echo "Error: npm no está instalado. Instala Node.js y npm antes de continuar." >&2
  exit 1
fi

if ! command -v curl >/dev/null 2>&1; then
  echo "Error: curl es necesario para comprobar la API y el frontend." >&2
  exit 1
fi

if ! curl --fail --silent --show-error --max-time 3 "${API_HEALTH_URL}" >/dev/null; then
  echo "Error: el API no responde en ${API_HEALTH_URL}. Inícialo desde backend/ y vuelve a ejecutar." >&2
  exit 1
fi

if [[ ! -d "${FRONTEND_DIR}/node_modules" ]]; then
  echo "Error: faltan las dependencias del frontend. Ejecuta 'npm install' dentro de frontend/." >&2
  exit 1
fi

if curl --fail --silent --max-time 1 "http://127.0.0.1:${FRONTEND_PORT}" >/dev/null; then
  echo "Error: ya hay un servicio usando el puerto ${FRONTEND_PORT}. Detén Vite antes de continuar." >&2
  exit 1
fi

echo "Iniciando frontend con API relativa (/api)..."
cd "${FRONTEND_DIR}"
VITE_API_BASE_URL=/api npm run dev -- --host 127.0.0.1 --strictPort >"${VITE_LOG}" 2>&1 &
VITE_PID=$!

frontend_ready=false
for _ in {1..30}; do
  if curl --fail --silent --max-time 1 "http://127.0.0.1:${FRONTEND_PORT}" >/dev/null; then
    frontend_ready=true
    break
  fi
  if ! kill -0 "${VITE_PID}" 2>/dev/null; then
    break
  fi
  sleep 1
done

if [[ "${frontend_ready}" != true ]]; then
  echo "Error: Vite no inició. Salida del frontend:" >&2
  cat "${VITE_LOG}" >&2
  exit 1
fi

echo "Frontend disponible en http://127.0.0.1:${FRONTEND_PORT}"
echo "ngrok expondrá el frontend. Las llamadas /api irán al backend local en el puerto 8080."
echo "Presiona Ctrl+C para cerrar el túnel y detener Vite."
ngrok http "${FRONTEND_PORT}"
