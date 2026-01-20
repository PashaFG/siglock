siglock sign [args...]   # обёртка над codesign
siglock status           # возвращает "free" или "blocked (PID)"
siglock unlock           # принудительно разблокировать

Коды возврата:
0 — успех
1 — заблокировано
2 — ошибка (например, не удалось запустить codesign)
