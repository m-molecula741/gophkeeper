-- Удаляем таблицы в обратном порядке (сначала дочернюю)
DROP TABLE IF EXISTS secrets;
DROP TABLE IF EXISTS users;

-- Расширение не удаляем — оно может использоваться другими БД