// Wishlist.java
import java.io.*;
import java.nio.file.*;
import java.time.LocalDate;
import java.util.*;
import java.util.stream.Collectors;

record Wish(int id, String title, String description, String category, int priority, Double price, String link, boolean fulfilled, String addedDate) implements Serializable {}

class WishesData implements Serializable {
    private static final long serialVersionUID = 1L;
    public List<Wish> wishes;
}

class Wishlist implements Serializable {
    private static final long serialVersionUID = 1L;
    private List<Wish> wishes = new ArrayList<>();
    private int nextId = 1;

    public Wish addWish(String title, String description, String category, int priority, Double price, String link, boolean fulfilled) {
        if (title.isBlank() || category.isBlank()) throw new IllegalArgumentException("Название и категория не могут быть пустыми");
        if (priority < 1 || priority > 3) throw new IllegalArgumentException("Приоритет должен быть 1, 2 или 3");
        if (price != null && price < 0) throw new IllegalArgumentException("Цена не может быть отрицательной");
        Wish wish = new Wish(nextId, title, description, category, priority, price, link, fulfilled, LocalDate.now().toString());
        wishes.add(wish);
        nextId++;
        return wish;
    }

    public Optional<Wish> findWish(int id) {
        return wishes.stream().filter(w -> w.id() == id).findFirst();
    }

    public boolean editWish(int id, Map<String, Object> updates) {
        Optional<Wish> opt = findWish(id);
        if (opt.isEmpty()) return false;
        Wish old = opt.get();
        wishes.remove(old);
        String title = (String) updates.getOrDefault("title", old.title());
        String description = (String) updates.getOrDefault("description", old.description());
        String category = (String) updates.getOrDefault("category", old.category());
        int priority = (int) updates.getOrDefault("priority", old.priority());
        Double price = (Double) updates.getOrDefault("price", old.price());
        String link = (String) updates.getOrDefault("link", old.link());
        boolean fulfilled = (boolean) updates.getOrDefault("fulfilled", old.fulfilled());
        Wish updated = new Wish(old.id(), title, description, category, priority, price, link, fulfilled, old.addedDate());
        wishes.add(updated);
        return true;
    }

    public boolean deleteWish(int id) {
        return wishes.removeIf(w -> w.id() == id);
    }

    public boolean toggleFulfilled(int id) {
        Optional<Wish> opt = findWish(id);
        if (opt.isEmpty()) return false;
        Wish old = opt.get();
        wishes.remove(old);
        Wish updated = new Wish(old.id(), old.title(), old.description(), old.category(), old.priority(), old.price(), old.link(), !old.fulfilled(), old.addedDate());
        wishes.add(updated);
        return true;
    }

    public List<Wish> searchWishes(String query) {
        String q = query.toLowerCase();
        return wishes.stream()
                .filter(w -> w.title().toLowerCase().contains(q) || w.description().toLowerCase().contains(q))
                .collect(Collectors.toList());
    }

    public List<Wish> filterByFulfilled(boolean fulfilled) {
        return wishes.stream().filter(w -> w.fulfilled() == fulfilled).collect(Collectors.toList());
    }

    public List<Wish> filterByCategory(String category) {
        return wishes.stream().filter(w -> w.category().equalsIgnoreCase(category)).collect(Collectors.toList());
    }

    public List<Wish> filterByPriority(int priority) {
        return wishes.stream().filter(w -> w.priority() == priority).collect(Collectors.toList());
    }

    public List<Wish> sortByPriority(boolean reverse) {
        return wishes.stream()
                .sorted((a, b) -> reverse ? Integer.compare(b.priority(), a.priority()) : Integer.compare(a.priority(), b.priority()))
                .collect(Collectors.toList());
    }

    public List<Wish> sortByPrice(boolean reverse) {
        return wishes.stream()
                .sorted((a, b) -> {
                    double pa = a.price() != null ? a.price() : 0;
                    double pb = b.price() != null ? b.price() : 0;
                    return reverse ? Double.compare(pb, pa) : Double.compare(pa, pb);
                })
                .collect(Collectors.toList());
    }

    public Map<String, Object> getStats() {
        int total = wishes.size();
        int fulfilled = filterByFulfilled(true).size();
        int unfulfilled = total - fulfilled;
        double avgPrice = wishes.stream().filter(w -> w.price() != null).mapToDouble(Wish::price).average().orElse(0);
        Map<String, Integer> categories = new HashMap<>();
        Map<Integer, Integer> priorities = new HashMap<>();
        priorities.put(1, 0);
        priorities.put(2, 0);
        priorities.put(3, 0);
        wishes.forEach(w -> {
            categories.put(w.category(), categories.getOrDefault(w.category(), 0) + 1);
            priorities.put(w.priority(), priorities.get(w.priority()) + 1);
        });
        Map<String, Object> stats = new HashMap<>();
        stats.put("total", total);
        stats.put("fulfilled", fulfilled);
        stats.put("unfulfilled", unfulfilled);
        stats.put("avg_price", avgPrice);
        stats.put("categories", categories);
        stats.put("priorities", priorities);
        return stats;
    }

    public void saveToFile(String filename) throws IOException {
        WishesData data = new WishesData();
        data.wishes = new ArrayList<>(wishes);
        try (ObjectOutputStream oos = new ObjectOutputStream(Files.newOutputStream(Paths.get(filename)))) {
            oos.writeObject(data);
        }
    }

    public void loadFromFile(String filename) throws IOException, ClassNotFoundException {
        try (ObjectInputStream ois = new ObjectInputStream(Files.newInputStream(Paths.get(filename)))) {
            WishesData data = (WishesData) ois.readObject();
            wishes = new ArrayList<>(data.wishes);
            for (Wish w : wishes) {
                if (w.id() >= nextId) nextId = w.id() + 1;
            }
        }
    }

    public void exportCSV(String filename) throws IOException {
        try (PrintWriter pw = new PrintWriter(Files.newBufferedWriter(Paths.get(filename)))) {
            pw.println("ID;Название;Описание;Категория;Приоритет;Цена;Ссылка;Исполнено;Дата добавления");
            for (Wish w : wishes) {
                pw.printf("%d;%s;%s;%s;%d;%s;%s;%s;%s%n",
                        w.id(), w.title(), w.description(), w.category(), w.priority(),
                        w.price() != null ? String.format("%.2f", w.price()) : "",
                        w.link(), w.fulfilled() ? "Да" : "Нет", w.addedDate());
            }
        }
    }

    public void importCSV(String filename) throws IOException {
        try (BufferedReader br = Files.newBufferedReader(Paths.get(filename))) {
            String line = br.readLine(); // header
            while ((line = br.readLine()) != null) {
                String[] parts = line.split(";");
                if (parts.length < 9) continue;
                String title = parts[1];
                String description = parts[2];
                String category = parts[3];
                int priority = Integer.parseInt(parts[4]);
                Double price = parts[5].isEmpty() ? null : Double.parseDouble(parts[5]);
                String link = parts[6];
                boolean fulfilled = parts[7].equals("Да");
                try {
                    addWish(title, description, category, priority, price, link, fulfilled);
                } catch (Exception e) {
                    System.out.println("Ошибка импорта строки: " + e.getMessage());
                }
            }
        }
    }

    public List<Wish> getWishes() { return Collections.unmodifiableList(wishes); }
}

public class WishlistApp {
    private static final Scanner scanner = new Scanner(System.in);

    private static String readString(String prompt) {
        System.out.print(prompt);
        return scanner.nextLine().trim();
    }

    private static int readInt(String prompt) {
        while (true) {
            try {
                System.out.print(prompt);
                return Integer.parseInt(scanner.nextLine().trim());
            } catch (NumberFormatException e) {
                System.out.println("Введите число.");
            }
        }
    }

    private static Double readDouble(String prompt) {
        while (true) {
            String input = readString(prompt);
            if (input.isEmpty()) return null;
            try {
                return Double.parseDouble(input);
            } catch (NumberFormatException e) {
                System.out.println("Введите число или оставьте пустым.");
            }
        }
    }

    private static boolean readBool(String prompt) {
        while (true) {
            String input = readString(prompt);
            if (input.equals("1")) return true;
            if (input.equals("0")) return false;
            System.out.println("Введите 1 или 0.");
        }
    }

    private static void printWish(Wish wish) {
        String status = wish.fulfilled() ? "✅ Исполнено" : "⏳ Желаемое";
        String priorityText = switch (wish.priority()) {
            case 1 -> "Низкий";
            case 2 -> "Средний";
            case 3 -> "Высокий";
            default -> "";
        };
        System.out.printf("#%d - %s (%s приоритет)%n", wish.id(), wish.title(), priorityText);
        if (!wish.description().isBlank())
            System.out.printf("   Описание: %s%n", wish.description());
        System.out.printf("   Категория: %s%n", wish.category());
        if (wish.price() != null)
            System.out.printf("   Цена: %.2f%n", wish.price());
        if (!wish.link().isBlank())
            System.out.printf("   Ссылка: %s%n", wish.link());
        System.out.printf("   %s, Добавлен: %s%n", status, wish.addedDate());
    }

    public static void main(String[] args) {
        Wishlist wishlist = new Wishlist();
        try {
            wishlist.loadFromFile("wishes_data.ser");
        } catch (IOException | ClassNotFoundException e) {
            System.out.println("Не удалось загрузить данные.");
        }

        while (true) {
            System.out.println("\n===== ВИШЛИСТ (Java) =====");
            System.out.println("1. Добавить желание");
            System.out.println("2. Показать все желания");
            System.out.println("3. Показать неисполненные желания");
            System.out.println("4. Показать исполненные желания");
            System.out.println("5. Найти желания по названию");
            System.out.println("6. Отметить желание как исполненное");
            System.out.println("7. Редактировать желание");
            System.out.println("8. Удалить желание");
            System.out.println("9. Показать статистику");
            System.out.println("10. Сохранить в файл");
            System.out.println("11. Загрузить из файла");
            System.out.println("12. Экспорт в CSV");
            System.out.println("13. Импорт из CSV");
            System.out.println("0. Выход");
            String choice = readString("Выберите действие: ");

            switch (choice) {
                case "0" -> { return; }
                case "1" -> {
                    String title = readString("Название: ");
                    if (title.isBlank()) { System.out.println("Название не может быть пустым."); continue; }
                    String description = readString("Описание (необязательно): ");
                    String category = readString("Категория: ");
                    if (category.isBlank()) { System.out.println("Категория не может быть пустой."); continue; }
                    int priority = readInt("Приоритет (1-низкий, 2-средний, 3-высокий): ");
                    Double price = readDouble("Цена (необязательно, число): ");
                    String link = readString("Ссылка (необязательно): ");
                    try {
                        Wish wish = wishlist.addWish(title, description, category, priority, price, link, false);
                        System.out.println("Желание добавлено с ID " + wish.id());
                    } catch (Exception e) {
                        System.out.println("Ошибка: " + e.getMessage());
                    }
                }
                case "2" -> {
                    if (wishlist.getWishes().isEmpty()) System.out.println("Нет желаний.");
                    else wishlist.getWishes().forEach(WishlistApp::printWish);
                }
                case "3" -> {
                    var unfulfilled = wishlist.filterByFulfilled(false);
                    if (unfulfilled.isEmpty()) System.out.println("Нет неисполненных желаний.");
                    else unfulfilled.forEach(WishlistApp::printWish);
                }
                case "4" -> {
                    var fulfilled = wishlist.filterByFulfilled(true);
                    if (fulfilled.isEmpty()) System.out.println("Нет исполненных желаний.");
                    else fulfilled.forEach(WishlistApp::printWish);
                }
                case "5" -> {
                    String query = readString("Введите часть названия или описания: ");
                    var results = wishlist.searchWishes(query);
                    if (results.isEmpty()) System.out.println("Желания не найдены.");
                    else results.forEach(WishlistApp::printWish);
                }
                case "6" -> {
                    int id = readInt("Введите ID желания: ");
                    if (wishlist.toggleFulfilled(id)) {
                        System.out.println("Статус желания изменён.");
                    } else {
                        System.out.println("Желание не найдено.");
                    }
                }
                case "7" -> {
                    int id = readInt("Введите ID желания для редактирования: ");
                    var opt = wishlist.findWish(id);
                    if (opt.isEmpty()) { System.out.println("Желание не найдено."); continue; }
                    Wish old = opt.get();
                    System.out.println("Оставьте поле пустым, чтобы не менять.");
                    String newTitle = readString("Название (" + old.title() + "): ");
                    String newDesc = readString("Описание (" + old.description() + "): ");
                    String newCat = readString("Категория (" + old.category() + "): ");
                    String newPriorityStr = readString("Приоритет (1-3) сейчас: " + old.priority() + ": ");
                    String newPriceStr = readString("Цена (" + (old.price() != null ? String.format("%.2f", old.price()) : "") + "): ");
                    String newLink = readString("Ссылка (" + old.link() + "): ");
                    String newFulfilledStr = readString("Статус (1-исполнено, 0-нет) сейчас: " + (old.fulfilled() ? "1" : "0") + ": ");
                    Map<String, Object> updates = new HashMap<>();
                    if (!newTitle.isBlank()) updates.put("title", newTitle);
                    if (!newDesc.isBlank()) updates.put("description", newDesc);
                    if (!newCat.isBlank()) updates.put("category", newCat);
                    if (!newPriorityStr.isBlank()) {
                        try { updates.put("priority", Integer.parseInt(newPriorityStr)); }
                        catch (NumberFormatException e) { System.out.println("Приоритет должен быть числом, пропускаем."); }
                    }
                    if (!newPriceStr.isBlank()) {
                        try { updates.put("price", Double.parseDouble(newPriceStr)); }
                        catch (NumberFormatException e) { System.out.println("Цена должна быть числом, пропускаем."); }
                    }
                    if (!newLink.isBlank()) updates.put("link", newLink);
                    if (!newFulfilledStr.isBlank()) updates.put("fulfilled", newFulfilledStr.equals("1"));
                    if (wishlist.editWish(id, updates)) System.out.println("Желание обновлено.");
                    else System.out.println("Ошибка обновления.");
                }
                case "8" -> {
                    int id = readInt("Введите ID желания для удаления: ");
                    if (wishlist.deleteWish(id)) System.out.println("Желание удалено.");
                    else System.out.println("Желание не найдено.");
                }
                case "9" -> {
                    var stats = wishlist.getStats();
                    System.out.println("\n=== СТАТИСТИКА ===");
                    System.out.println("Всего желаний: " + stats.get("total"));
                    System.out.println("Исполнено: " + stats.get("fulfilled"));
                    System.out.println("Не исполнено: " + stats.get("unfulfilled"));
                    System.out.printf("Средняя цена: %.2f%n", stats.get("avg_price"));
                    System.out.println("По категориям:");
                    @SuppressWarnings("unchecked")
                    Map<String, Integer> categories = (Map<String, Integer>) stats.get("categories");
                    categories.forEach((cat, count) -> System.out.println("  " + cat + ": " + count));
                    System.out.println("По приоритетам:");
                    @SuppressWarnings("unchecked")
                    Map<Integer, Integer> priorities = (Map<Integer, Integer>) stats.get("priorities");
                    priorities.forEach((p, count) -> {
                        String name = switch (p) { case 1 -> "Низкий"; case 2 -> "Средний"; case 3 -> "Высокий"; default -> ""; };
                        System.out.println("  " + name + ": " + count);
                    });
                }
                case "10" -> {
                    try {
                        wishlist.saveToFile("wishes_data.ser");
                        System.out.println("Сохранено.");
                    } catch (IOException e) {
                        System.out.println("Ошибка сохранения: " + e.getMessage());
                    }
                }
                case "11" -> {
                    try {
                        wishlist.loadFromFile("wishes_data.ser");
                        System.out.println("Загружено.");
                    } catch (IOException | ClassNotFoundException e) {
                        System.out.println("Ошибка загрузки: " + e.getMessage());
                    }
                }
                case "12" -> {
                    try {
                        wishlist.exportCSV("wishes_export.csv");
                        System.out.println("Экспортировано в wishes_export.csv");
                    } catch (IOException e) {
                        System.out.println("Ошибка экспорта: " + e.getMessage());
                    }
                }
                case "13" -> {
                    try {
                        wishlist.importCSV("wishes_export.csv");
                        System.out.println("Импортировано из wishes_export.csv");
                    } catch (IOException e) {
                        System.out.println("Ошибка импорта: " + e.getMessage());
                    }
                }
                default -> System.out.println("Неизвестная команда.");
            }
        }
    }
}
