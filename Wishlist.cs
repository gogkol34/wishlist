// Wishlist.cs
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text.Json;
using System.Text.Json.Serialization;

public record Wish(
    int Id,
    string Title,
    string Description,
    string Category,
    int Priority,
    double? Price,
    string Link,
    bool Fulfilled,
    string AddedDate
);

public class WishesData
{
    public List<Wish> Wishes { get; set; } = new();
}

public class Wishlist
{
    private List<Wish> wishes = new();
    private int nextId = 1;

    public IReadOnlyList<Wish> Wishes => wishes.AsReadOnly();

    public Wish AddWish(string title, string description, string category, int priority, double? price, string link, bool fulfilled = false)
    {
        if (string.IsNullOrWhiteSpace(title) || string.IsNullOrWhiteSpace(category))
            throw new ArgumentException("Название и категория не могут быть пустыми");
        if (priority < 1 || priority > 3)
            throw new ArgumentException("Приоритет должен быть 1, 2 или 3");
        if (price.HasValue && price.Value < 0)
            throw new ArgumentException("Цена не может быть отрицательной");
        var wish = new Wish(nextId, title, description ?? "", category, priority, price, link ?? "", fulfilled, DateTime.Now.ToString("yyyy-MM-dd"));
        wishes.Add(wish);
        nextId++;
        return wish;
    }

    public Wish? FindWish(int id) => wishes.FirstOrDefault(w => w.Id == id);

    public bool EditWish(int id, Dictionary<string, object> updates)
    {
        var old = FindWish(id);
        if (old == null) return false;
        wishes.Remove(old);
        string title = updates.ContainsKey("title") ? (string)updates["title"] : old.Title;
        string description = updates.ContainsKey("description") ? (string)updates["description"] : old.Description;
        string category = updates.ContainsKey("category") ? (string)updates["category"] : old.Category;
        int priority = updates.ContainsKey("priority") ? (int)updates["priority"] : old.Priority;
        double? price = updates.ContainsKey("price") ? (double?)updates["price"] : old.Price;
        string link = updates.ContainsKey("link") ? (string)updates["link"] : old.Link;
        bool fulfilled = updates.ContainsKey("fulfilled") ? (bool)updates["fulfilled"] : old.Fulfilled;
        var updated = new Wish(old.Id, title, description, category, priority, price, link, fulfilled, old.AddedDate);
        wishes.Add(updated);
        return true;
    }

    public bool DeleteWish(int id) => wishes.RemoveAll(w => w.Id == id) > 0;

    public bool ToggleFulfilled(int id)
    {
        var old = FindWish(id);
        if (old == null) return false;
        wishes.Remove(old);
        var updated = old with { Fulfilled = !old.Fulfilled };
        wishes.Add(updated);
        return true;
    }

    public List<Wish> SearchWishes(string query)
    {
        var q = query.ToLower();
        return wishes.Where(w => w.Title.ToLower().Contains(q) || w.Description.ToLower().Contains(q)).ToList();
    }

    public List<Wish> FilterByFulfilled(bool fulfilled) => wishes.Where(w => w.Fulfilled == fulfilled).ToList();

    public List<Wish> FilterByCategory(string category) =>
        wishes.Where(w => string.Equals(w.Category, category, StringComparison.OrdinalIgnoreCase)).ToList();

    public List<Wish> FilterByPriority(int priority) => wishes.Where(w => w.Priority == priority).ToList();

    public List<Wish> SortByPriority(bool reverse) =>
        wishes.OrderBy(w => w.Priority).Reverse(reverse).ToList();

    public List<Wish> SortByPrice(bool reverse) =>
        wishes.OrderBy(w => w.Price ?? 0).Reverse(reverse).ToList();

    public Dictionary<string, object> GetStats()
    {
        int total = wishes.Count;
        int fulfilled = FilterByFulfilled(true).Count;
        int unfulfilled = total - fulfilled;
        double avgPrice = wishes.Where(w => w.Price.HasValue).Average(w => w.Price.Value);
        var categories = wishes.GroupBy(w => w.Category).ToDictionary(g => g.Key, g => g.Count());
        var priorities = new Dictionary<int, int> { [1] = 0, [2] = 0, [3] = 0 };
        foreach (var w in wishes) priorities[w.Priority]++;
        return new Dictionary<string, object>
        {
            ["total"] = total,
            ["fulfilled"] = fulfilled,
            ["unfulfilled"] = unfulfilled,
            ["avg_price"] = avgPrice,
            ["categories"] = categories,
            ["priorities"] = priorities
        };
    }

    public void SaveToFile(string filename)
    {
        var data = new WishesData { Wishes = wishes };
        var options = new JsonSerializerOptions { WriteIndented = true };
        string json = JsonSerializer.Serialize(data, options);
        File.WriteAllText(filename, json);
    }

    public void LoadFromFile(string filename)
    {
        if (!File.Exists(filename)) return;
        string json = File.ReadAllText(filename);
        var data = JsonSerializer.Deserialize<WishesData>(json);
        if (data != null)
        {
            wishes = data.Wishes;
            nextId = wishes.Any() ? wishes.Max(w => w.Id) + 1 : 1;
        }
    }

    public void ExportCSV(string filename)
    {
        using var writer = new StreamWriter(filename);
        writer.WriteLine("ID;Название;Описание;Категория;Приоритет;Цена;Ссылка;Исполнено;Дата добавления");
        foreach (var w in wishes)
        {
            writer.WriteLine($"{w.Id};{w.Title};{w.Description};{w.Category};{w.Priority};{(w.Price.HasValue ? w.Price.Value.ToString("F2") : "")};{w.Link};{(w.Fulfilled ? "Да" : "Нет")};{w.AddedDate}");
        }
    }

    public void ImportCSV(string filename)
    {
        if (!File.Exists(filename)) throw new FileNotFoundException("Файл не найден");
        using var reader = new StreamReader(filename);
        string header = reader.ReadLine(); // skip header
        while (!reader.EndOfStream)
        {
            string line = reader.ReadLine();
            var parts = line.Split(';');
            if (parts.Length < 9) continue;
            string title = parts[1];
            string description = parts[2];
            string category = parts[3];
            int priority = int.Parse(parts[4]);
            double? price = string.IsNullOrEmpty(parts[5]) ? null : double.Parse(parts[5]);
            string link = parts[6];
            bool fulfilled = parts[7] == "Да";
            try
            {
                AddWish(title, description, category, priority, price, link, fulfilled);
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Ошибка импорта строки: {ex.Message}");
            }
        }
    }
}

public static class Program
{
    private static string ReadString(string prompt)
    {
        Console.Write(prompt);
        return Console.ReadLine()?.Trim() ?? "";
    }

    private static int ReadInt(string prompt)
    {
        while (true)
        {
            Console.Write(prompt);
            if (int.TryParse(Console.ReadLine(), out int result))
                return result;
            Console.WriteLine("Введите число.");
        }
    }

    private static double? ReadDouble(string prompt)
    {
        while (true)
        {
            string input = ReadString(prompt);
            if (string.IsNullOrEmpty(input)) return null;
            if (double.TryParse(input, out double result))
                return result;
            Console.WriteLine("Введите число или оставьте пустым.");
        }
    }

    private static bool ReadBool(string prompt)
    {
        while (true)
        {
            string input = ReadString(prompt);
            if (input == "1") return true;
            if (input == "0") return false;
            Console.WriteLine("Введите 1 или 0.");
        }
    }

    private static void PrintWish(Wish wish)
    {
        string status = wish.Fulfilled ? "✅ Исполнено" : "⏳ Желаемое";
        string priorityText = wish.Priority switch { 1 => "Низкий", 2 => "Средний", 3 => "Высокий", _ => "" };
        Console.WriteLine($"#{wish.Id} - {wish.Title} ({priorityText} приоритет)");
        if (!string.IsNullOrWhiteSpace(wish.Description))
            Console.WriteLine($"   Описание: {wish.Description}");
        Console.WriteLine($"   Категория: {wish.Category}");
        if (wish.Price.HasValue)
            Console.WriteLine($"   Цена: {wish.Price.Value:F2}");
        if (!string.IsNullOrWhiteSpace(wish.Link))
            Console.WriteLine($"   Ссылка: {wish.Link}");
        Console.WriteLine($"   {status}, Добавлен: {wish.AddedDate}");
    }

    public static void Main()
    {
        var wishlist = new Wishlist();
        try { wishlist.LoadFromFile("wishes_data.json"); }
        catch { Console.WriteLine("Не удалось загрузить данные."); }

        while (true)
        {
            Console.WriteLine("\n===== ВИШЛИСТ (C#) =====");
            Console.WriteLine("1. Добавить желание");
            Console.WriteLine("2. Показать все желания");
            Console.WriteLine("3. Показать неисполненные желания");
            Console.WriteLine("4. Показать исполненные желания");
            Console.WriteLine("5. Найти желания по названию");
            Console.WriteLine("6. Отметить желание как исполненное");
            Console.WriteLine("7. Редактировать желание");
            Console.WriteLine("8. Удалить желание");
            Console.WriteLine("9. Показать статистику");
            Console.WriteLine("10. Сохранить в файл");
            Console.WriteLine("11. Загрузить из файла");
            Console.WriteLine("12. Экспорт в CSV");
            Console.WriteLine("13. Импорт из CSV");
            Console.WriteLine("0. Выход");
            string choice = ReadString("Выберите действие: ");

            switch (choice)
            {
                case "0": return;
                case "1":
                    string title = ReadString("Название: ");
                    if (string.IsNullOrWhiteSpace(title)) { Console.WriteLine("Название не может быть пустым."); continue; }
                    string description = ReadString("Описание (необязательно): ");
                    string category = ReadString("Категория: ");
                    if (string.IsNullOrWhiteSpace(category)) { Console.WriteLine("Категория не может быть пустой."); continue; }
                    int priority = ReadInt("Приоритет (1-низкий, 2-средний, 3-высокий): ");
                    double? price = ReadDouble("Цена (необязательно, число): ");
                    string link = ReadString("Ссылка (необязательно): ");
                    try
                    {
                        var wish = wishlist.AddWish(title, description, category, priority, price, link);
                        Console.WriteLine($"Желание добавлено с ID {wish.Id}");
                    }
                    catch (Exception ex) { Console.WriteLine($"Ошибка: {ex.Message}"); }
                    break;
                case "2":
                    if (!wishlist.Wishes.Any()) Console.WriteLine("Нет желаний.");
                    else foreach (var w in wishlist.Wishes) PrintWish(w);
                    break;
                case "3":
                    var unfulfilled = wishlist.FilterByFulfilled(false);
                    if (!unfulfilled.Any()) Console.WriteLine("Нет неисполненных желаний.");
                    else foreach (var w in unfulfilled) PrintWish(w);
                    break;
                case "4":
                    var fulfilled = wishlist.FilterByFulfilled(true);
                    if (!fulfilled.Any()) Console.WriteLine("Нет исполненных желаний.");
                    else foreach (var w in fulfilled) PrintWish(w);
                    break;
                case "5":
                    string query = ReadString("Введите часть названия или описания: ");
                    var results = wishlist.SearchWishes(query);
                    if (!results.Any()) Console.WriteLine("Желания не найдены.");
                    else foreach (var w in results) PrintWish(w);
                    break;
                case "6":
                    int id = ReadInt("Введите ID желания: ");
                    if (wishlist.ToggleFulfilled(id)) Console.WriteLine("Статус желания изменён.");
                    else Console.WriteLine("Желание не найдено.");
                    break;
                case "7":
                    int eid = ReadInt("Введите ID желания для редактирования: ");
                    var old = wishlist.FindWish(eid);
                    if (old == null) { Console.WriteLine("Желание не найдено."); continue; }
                    Console.WriteLine("Оставьте поле пустым, чтобы не менять.");
                    string newTitle = ReadString($"Название ({old.Title}): ");
                    string newDesc = ReadString($"Описание ({old.Description}): ");
                    string newCat = ReadString($"Категория ({old.Category}): ");
                    string newPriorityStr = ReadString($"Приоритет (1-3) сейчас: {old.Priority}: ");
                    string newPriceStr = ReadString($"Цена ({old.Price}): ");
                    string newLink = ReadString($"Ссылка ({old.Link}): ");
                    string newFulfilledStr = ReadString($"Статус (1-исполнено, 0-нет) сейчас: {(old.Fulfilled ? "1" : "0")}: ");
                    var updates = new Dictionary<string, object>();
                    if (!string.IsNullOrWhiteSpace(newTitle)) updates["title"] = newTitle;
                    if (!string.IsNullOrWhiteSpace(newDesc)) updates["description"] = newDesc;
                    if (!string.IsNullOrWhiteSpace(newCat)) updates["category"] = newCat;
                    if (!string.IsNullOrWhiteSpace(newPriorityStr))
                    {
                        if (int.TryParse(newPriorityStr, out int p)) updates["priority"] = p;
                        else Console.WriteLine("Приоритет должен быть числом, пропускаем.");
                    }
                    if (!string.IsNullOrWhiteSpace(newPriceStr))
                    {
                        if (double.TryParse(newPriceStr, out double p)) updates["price"] = p;
                        else Console.WriteLine("Цена должна быть числом, пропускаем.");
                    }
                    if (!string.IsNullOrWhiteSpace(newLink)) updates["link"] = newLink;
                    if (!string.IsNullOrWhiteSpace(newFulfilledStr)) updates["fulfilled"] = newFulfilledStr == "1";
                    if (wishlist.EditWish(eid, updates)) Console.WriteLine("Желание обновлено.");
                    else Console.WriteLine("Ошибка обновления.");
                    break;
                case "8":
                    int delId = ReadInt("Введите ID желания для удаления: ");
                    if (wishlist.DeleteWish(delId)) Console.WriteLine("Желание удалено.");
                    else Console.WriteLine("Желание не найдено.");
                    break;
                case "9":
                    var stats = wishlist.GetStats();
                    Console.WriteLine("\n=== СТАТИСТИКА ===");
                    Console.WriteLine($"Всего желаний: {stats["total"]}");
                    Console.WriteLine($"Исполнено: {stats["fulfilled"]}");
                    Console.WriteLine($"Не исполнено: {stats["unfulfilled"]}");
                    Console.WriteLine($"Средняя цена: {stats["avg_price"]:F2}");
                    Console.WriteLine("По категориям:");
                    var categories = (Dictionary<string, int>)stats["categories"];
                    foreach (var kv in categories) Console.WriteLine($"  {kv.Key}: {kv.Value}");
                    Console.WriteLine("По приоритетам:");
                    var priorities = (Dictionary<int, int>)stats["priorities"];
                    foreach (var kv in priorities)
                    {
                        string name = kv.Key switch { 1 => "Низкий", 2 => "Средний", 3 => "Высокий", _ => "" };
                        Console.WriteLine($"  {name}: {kv.Value}");
                    }
                    break;
                case "10":
                    try { wishlist.SaveToFile("wishes_data.json"); Console.WriteLine("Сохранено."); }
                    catch (Exception ex) { Console.WriteLine($"Ошибка: {ex.Message}"); }
                    break;
                case "11":
                    try { wishlist.LoadFromFile("wishes_data.json"); Console.WriteLine("Загружено."); }
                    catch (Exception ex) { Console.WriteLine($"Ошибка: {ex.Message}"); }
                    break;
                case "12":
                    try { wishlist.ExportCSV("wishes_export.csv"); Console.WriteLine("Экспортировано в wishes_export.csv"); }
                    catch (Exception ex) { Console.WriteLine($"Ошибка: {ex.Message}"); }
                    break;
                case "13":
                    try { wishlist.ImportCSV("wishes_export.csv"); Console.WriteLine("Импортировано из wishes_export.csv"); }
                    catch (Exception ex) { Console.WriteLine($"Ошибка: {ex.Message}"); }
                    break;
                default: Console.WriteLine("Неизвестная команда."); break;
            }
        }
    }
}
