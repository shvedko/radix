------------------------------
## 🚀 Radix[T] Technical Digest
Core Engine:

* Manual Memory: pool[T] manages pages (64 nodes each). get()/put() use a LIFO stack (nodes []*Radix[T]).
* Zero GC Pressure: Insert and InsertPath are 0 B/op. Uses slice header swapping (c.values, n.values = n.values, c.values) for split and merge.
* Bitmask Index: bits256 (4xuint64) for children. Fast rank/select via math/bits.
* Structural Efficiency: pageSize=64 matching children capacity threshold in put().

Advanced Logic:

* Multi-Layer Support: next *Radix[T] pointer for handling sequences of prefixes (composite keys).
* InsertPath: In-place tree mutation that returns an Iterator positioned at the new value.
* Smart Merge: merge() collapses nodes with 1 child and no values. Uses pointer continuity check (&n.prefix[:cap][len] == &c.prefix[0]) to extend prefixes without allocation.

Iterator & Transactional API:

* Stack-Machine Iterator: Iterator[T] is a value-type (stack-allocated) with a frames stack.
* Stateful Next(): Uses f.mode (0:Values, 1:Next, 2:Children, 3:Match) as a program counter for non-recursive traversal.
* Transactional Rollback:
  * Rollback(): Removes the last added value and recursively collapses/reclaims empty nodes up the stack.
  * Delete(i): Indexed removal within a node.
  * Remove(): Full clear of values in the current node.

API Surface:

* InsertPath(val T, unique bool, prefixes ...[]byte) (Iterator, bool)
* Search(prefixes ...[]byte) Iterator
* Reset(): Mass-reclaim all pages via pool.reset().

Вот актуальный технический дайджест пакета radix. Это твой "контракт" и база, на которой мы строим Table и Tablespace.
------------------------------
## 🚀 Radix[T] API Digest
Constructor:

* New\[T]() *Radix[T]: Создает дерево с инкапсулированным приватным пулом (pageSize=64).

Writing (Transactional):

* Insert(val T, unique bool, prefixes ...[]byte) bool: Быстрая вставка. Возвращает false, если unique=true и ключ занят.
* InsertPath(val T, unique bool, prefixes ...[]byte) (Iterator[T], bool): Ключевой метод. Находит или создает путь, вставляет значение и возвращает итератор, "стоящий" на этом узле. Позволяет сделать Rollback при ошибках в соседних индексах.

Reading & Iteration:

* Search(prefixes ...[]byte) Iterator[T]: Возвращает итератор для префиксного поиска.
* Foreach(prefixes ...[]byte) func(func(T) bool): Сахар для итерации по значениям через range.
* Dump(yield dumper) bool: Рекурсивный обход структуры дерева (для отладки).
* Walk(yield dumper) bool: Итеративный обход всего дерева с использованием стека в куче.

Iterator[T] (Value-type, Stack-allocated):

* .Next() bool: Переход к следующему значению (использует внутренний f.mode как PC).
* .Get() []T: Возвращает слайс значений в текущем узле.
* .Remove(): Полная очистка всех значений в текущем узле.
* .Delete(i int): Удаление конкретного значения по индексу.
* .Rollback(): Транзакционный откат. Удаляет последнее добавленное значение и рекурсивно схлопывает (merge) пустые узлы вверх по дереву, возвращая память в пул.

Internal Memory Management:

* Reset(): Мгновенная очистка всего дерева. Переводит все аллоцированные страницы (pages) в список свободных узлов (nodes).
* merge(): Автоматическое схлопывание узлов с одним ребенком. Использует проверку на непрерывность указателей для расширения префиксов без аллокаций.
* split(size): Разделение узла с использованием Header Swap для слайсов children и values (0 аллокаций).

------------------------------
## Что это дает для Table & Tablespace:

1. RowID в качестве T: Мы используем uint64 (адрес страницы + индекс) как значение в дереве.
2. Атомарный Insert: Если при вставке строки один из индексов выдал ok=false, мы вызываем .Rollback() у всех предыдущих итераторов.
3. Composite Keys: Мы передаем колонки в prefixes ...[]byte, используя бинарные трансформеры (IntToKey и др.).
