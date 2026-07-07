// wishlist.cpp
#include <iostream>
#include <vector>
#include <string>
#include <fstream>
#include <sstream>
#include <algorithm>
#include <iomanip>
#include <ctime>
#include <map>
#include <variant>
#include <regex>
#include <cctype>
#include <random>

using namespace std;

struct Wish {
    int id;
    string title;
    string description;
    string category;
    int priority;
    double price;
    bool hasPrice;
    string link;
    bool fulfilled;
    string addedDate;

    Wish(int id, const string& title, const string& description, const string& category,
         int priority, double price, bool hasPrice, const string& link, bool fulfilled, const string& addedDate = "")
        : id(id), title(title), description(description), category(category), priority(priority),
          price(price), hasPrice(hasPrice), link(link), fulfilled(fulfilled), addedDate(addedDate) {
        if (addedDate.empty()) {
            time_t now = time(nullptr);
            tm* tm_now = localtime(&now);
            char buf[11];
            strftime(buf, sizeof(buf), "%Y-%m-%d", tm_now);
            this->addedDate = string(buf);
        }
    }
};

class Wishlist {
private:
    vector<Wish> wishes;
    int nextId = 1;

public:
    Wish addWish(const string& title, const string& description, const string& category,
                 int priority, double price, bool hasPrice, const string& link, bool fulfilled = false) {
        if (title.empty() || category.empty()) throw invalid_argument("Название и категория не могут быть пустыми");
        if (priority < 1 || priority > 3) throw invalid_argument("Приоритет должен быть 1, 2 или 3");
        if (hasPrice && price < 0) throw invalid_argument("Цена не может быть отрицательной");
        Wish wish(nextId, title, description, category, priority, price, hasPrice, link, fulfilled);
        wishes.push_back(wish);
        nextId++;
        return wish;
    }

    Wish* findWish(int id) {
        auto it = find_if(wishes.begin(), wishes.end(), [id](const Wish& w) { return w.id == id; });
        return it != wishes.end() ? &(*it) : nullptr;
    }

    bool editWish(int id, const map<string, string>& updates) {
        Wish* wish = findWish(id);
        if (!wish) return false;
        for (const auto& [key, value] : updates) {
            if (key == "title") wish->title = value;
            else if (key == "description") wish->description = value;
            else if (key == "category") wish->category = value;
            else if (key == "priority") wish->priority = stoi(value);
            else if (key == "price") {
                if (value.empty()) { wish->hasPrice = false; }
                else { wish->price = stod(value); wish->hasPrice = true; }
            }
            else if (key == "link") wish->link = value;
            else if (key == "fulfilled") wish->fulfilled = (value == "1");
        }
        return true;
    }

    bool deleteWish(int id) {
        auto it = find_if(wishes.begin(), wishes.end(), [id](const Wish& w) { return w.id == id; });
        if (it == wishes.end()) return false;
        wishes.erase(it);
        return true;
    }

    bool toggleFulfilled(int id) {
        Wish* wish = findWish(id);
        if (!wish) return false;
        wish->fulfilled = !wish->fulfilled;
        return true;
    }

    vector<Wish> searchWishes(const string& query) {
        string q = query;
        transform(q.begin(), q.end(), q.begin(), ::tolower);
        vector<Wish> result;
        for (const auto& w : wishes) {
            string titleLower = w.title, descLower = w.description;
            transform(titleLower.begin(), titleLower.end(), titleLower.begin(), ::tolower);
            transform(descLower.begin(), descLower.end(), descLower.begin(), ::tolower);
            if (titleLower.find(q) != string::npos || descLower.find(q) != string::npos) {
                result.push_back(w);
            }
        }
        return result;
    }

    vector<Wish> filterByFulfilled(bool fulfilled) const {
        vector<Wish> result;
        for (const auto& w : wishes) {
            if (w.fulfilled == fulfilled) result.push_back(w);
        }
        return result;
    }

    vector<Wish> filterByCategory(const string& category) const {
        vector<Wish> result;
        for (const auto& w : wishes) {
            if (w.category == category) result.push_back(w);
        }
        return result;
    }

    vector<Wish> filterByPriority(int priority) const {
        vector<Wish> result;
        for (const auto& w : wishes) {
            if (w.priority == priority) result.push_back(w);
        }
        return result;
    }

    vector<Wish> sortByPriority(bool reverse = true) {
        vector<Wish> sorted = wishes;
        sort(sorted.begin(), sorted.end(), [reverse](const Wish& a, const Wish& b) {
            return reverse ? a.priority > b.priority : a.priority < b.priority;
        });
        return sorted;
    }

    vector<Wish> sortByPrice(bool reverse = false) {
        vector<Wish> sorted = wishes;
        sort(sorted.begin(), sorted.end(), [reverse](const Wish& a, const Wish& b) {
            double pa = a.hasPrice ? a.price : 0;
            double pb = b.hasPrice ? b.price : 0;
            return reverse ? pa > pb : pa < pb;
        });
        return sorted;
    }

    map<string, variant<int, double, map<string, int>>> getStats() const {
        int total = wishes.size();
        int fulfilled = filterByFulfilled(true).size();
        int unfulfilled = total - fulfilled;
        double sumPrice = 0;
        int priceCount = 0;
        map<string, int> categories;
        map<int, int> priorities = {{1,0},{2,0},{3,0}};
        for (const auto& w : wishes) {
            if (w.hasPrice) { sumPrice += w.price; priceCount++; }
            categories[w.category]++;
            priorities[w.priority]++;
        }
        double avgPrice = priceCount > 0 ? sumPrice / priceCount : 0.0;
        map<string, variant<int, double, map<string, int>>> stats;
        stats["total"] = total;
        stats["fulfilled"] = fulfilled;
        stats["unfulfilled"] = unfulfilled;
        stats["avg_price"] = avgPrice;
        stats["categories"] = categories;
        stats["priorities"] = priorities;
        return stats;
    }

    void saveToFile(const string& filename = "wishes_data.txt") {
        ofstream out(filename);
        if (!out) return;
        for (const auto& w : wishes) {
            out << w.id << '|'
                << w.title << '|'
                << w.description << '|'
                << w.category << '|'
                << w.priority << '|'
                << (w.hasPrice ? to_string(w.price) : "") << '|'
                << w.link << '|'
                << w.fulfilled << '|'
                << w.addedDate << '\n';
        }
    }

    void loadFromFile(const string& filename = "wishes_data.txt") {
        ifstream in(filename);
        if (!in) return;
        wishes.clear();
        string line;
        while (getline(in, line)) {
            stringstream ss(line);
            string idStr, title, description, category, priorityStr, priceStr, link, fulfilledStr, addedDate;
            getline(ss, idStr, '|');
            getline(ss, title, '|');
            getline(ss, description, '|');
            getline(ss, category, '|');
            getline(ss, priorityStr, '|');
            getline(ss, priceStr, '|');
            getline(ss, link, '|');
            getline(ss, fulfilledStr, '|');
            getline(ss, addedDate, '|');
            int id = stoi(idStr);
            int priority = stoi(priorityStr);
            bool hasPrice = !priceStr.empty();
            double price = hasPrice ? stod(priceStr) : 0.0;
            bool fulfilled = (fulfilledStr == "1");
            wishes.emplace_back(id, title, description, category, priority, price, hasPrice, link, fulfilled, addedDate);
            if (id >= nextId) nextId = id + 1;
        }
    }

    void exportCSV(const string& filename = "wishes_export.csv") {
        ofstream out(filename);
        if (!out) return;
        out << "ID;Название;Описание;Категория;Приоритет;Цена;Ссылка;Исполнено;Дата добавления\n";
        for (const auto& w : wishes) {
            out << w.id << ';'
                << w.title << ';'
                << w.description << ';'
                << w.category << ';'
                << w.priority << ';'
                << (w.hasPrice ? to_string(w.price) : "") << ';'
                << w.link << ';'
                << (w.fulfilled ? "Да" : "Нет") << ';'
                << w.addedDate << '\n';
        }
    }

    void importCSV(const string& filename = "wishes_export.csv") {
        ifstream in(filename);
        if (!in) return;
        string line;
        getline(in, line); // header
        while (getline(in, line)) {
            stringstream ss(line);
            string idStr, title, description, category, priorityStr, priceStr, link, fulfilledStr, addedDate;
            getline(ss, idStr, ';');
            getline(ss, title, ';');
            getline(ss, description, ';');
            getline(ss, category, ';');
            getline(ss, priorityStr, ';');
            getline(ss, priceStr, ';');
            getline(ss, link, ';');
            getline(ss, fulfilledStr, ';');
            getline(ss, addedDate, ';');
            try {
                int priority = stoi(priorityStr);
                bool hasPrice = !priceStr.empty();
                double price = hasPrice ? stod(priceStr) : 0.0;
                bool fulfilled = (fulfilledStr == "Да");
                addWish(title, description, category, priority, price, hasPrice, link, fulfilled);
            } catch (const exception& e) {
                cout << "Ошибка импорта строки: " << e.what() << "\n";
            }
        }
    }

    const vector<Wish>& getWishes() const { return wishes; }
};

string readString(const string& prompt) {
    cout << prompt;
    string input;
    getline(cin, input);
    return input;
}

int readInt(const string& prompt) {
    while (true) {
        cout << prompt;
        string input;
        getline(cin, input);
        try {
            return stoi(input);
        } catch (...) {
            cout << "Введите число.\n";
        }
    }
}

double readDouble(const string& prompt) {
    while (true) {
        string input = readString(prompt);
        if (input.empty()) return 0;
        try {
            return stod(input);
        } catch (...) {
            cout << "Введите число или оставьте пустым.\n";
        }
    }
}

bool hasPriceInput(const string& prompt) {
    string input = readString(prompt);
    return !input.empty();
}

bool readBool(const string& prompt) {
    while (true) {
        string input = readString(prompt);
        if (input == "1") return true;
        if (input == "0") return false;
        cout << "Введите 1 или 0.\n";
    }
}

void printWish(const Wish& wish) {
    string status = wish.fulfilled ? "✅ Исполнено" : "⏳ Желаемое";
    string priorityText;
    switch (wish.priority) {
        case 1: priorityText = "Низкий"; break;
        case 2: priorityText = "Средний"; break;
        case 3: priorityText = "Высокий"; break;
        default: priorityText = "";
    }
    cout << "#" << wish.id << " - " << wish.title << " (" << priorityText << " приоритет)\n";
    if (!wish.description.empty()) cout << "   Описание: " << wish.description << "\n";
    cout << "   Категория: " << wish.category << "\n";
    if (wish.hasPrice) cout << "   Цена: " << fixed << setprecision(2) << wish.price << "\n";
    if (!wish.link.empty()) cout << "   Ссылка: " << wish.link << "\n";
    cout << "   " << status << ", Добавлен: " << wish.addedDate << "\n";
}

int main() {
    Wishlist wishlist;
    wishlist.loadFromFile();

    while (true) {
        cout << "\n===== ВИШЛИСТ (C++) =====" << endl;
        cout << "1. Добавить желание\n";
        cout << "2. Показать все желания\n";
        cout << "3. Показать неисполненные желания\n";
        cout << "4. Показать исполненные желания\n";
        cout << "5. Найти желания по названию\n";
        cout << "6. Отметить желание как исполненное\n";
        cout << "7. Редактировать желание\n";
        cout << "8. Удалить желание\n";
        cout << "9. Показать статистику\n";
        cout << "10. Сохранить в файл\n";
        cout << "11. Загрузить из файла\n";
        cout << "12. Экспорт в CSV\n";
        cout << "13. Импорт из CSV\n";
        cout << "0. Выход\n";
        string choice = readString("Выберите действие: ");

        if (choice == "0") break;

        if (choice == "1") {
            string title = readString("Название: ");
            if (title.empty()) { cout << "Название не может быть пустым.\n"; continue; }
            string description = readString("Описание (необязательно): ");
            string category = readString("Категория: ");
            if (category.empty()) { cout << "Категория не может быть пустой.\n"; continue; }
            int priority = readInt("Приоритет (1-низкий, 2-средний, 3-высокий): ");
            bool hasPrice = false;
            double price = 0;
            string priceStr = readString("Цена (необязательно, число): ");
            if (!priceStr.empty()) {
                try { price = stod(priceStr); hasPrice = true; }
                catch (...) { cout << "Некорректное число, цена не сохранена.\n"; }
            }
            string link = readString("Ссылка (необязательно): ");
            try {
                Wish wish = wishlist.addWish(title, description, category, priority, price, hasPrice, link);
                cout << "Желание добавлено с ID " << wish.id << "\n";
            } catch (const exception& e) {
                cout << "Ошибка: " << e.what() << "\n";
            }
        } else if (choice == "2") {
            if (wishlist.getWishes().empty()) {
                cout << "Нет желаний.\n";
            } else {
                for (const auto& w : wishlist.getWishes()) printWish(w);
            }
        } else if (choice == "3") {
            auto unfulfilled = wishlist.filterByFulfilled(false);
            if (unfulfilled.empty()) cout << "Нет неисполненных желаний.\n";
            else for (const auto& w : unfulfilled) printWish(w);
        } else if (choice == "4") {
            auto fulfilled = wishlist.filterByFulfilled(true);
            if (fulfilled.empty()) cout << "Нет исполненных желаний.\n";
            else for (const auto& w : fulfilled) printWish(w);
        } else if (choice == "5") {
            string query = readString("Введите часть названия или описания: ");
            auto results = wishlist.searchWishes(query);
            if (results.empty()) cout << "Желания не найдены.\n";
            else for (const auto& w : results) printWish(w);
        } else if (choice == "6") {
            int id = readInt("Введите ID желания: ");
            if (wishlist.toggleFulfilled(id)) {
                cout << "Статус желания изменён.\n";
            } else {
                cout << "Желание не найдено.\n";
            }
        } else if (choice == "7") {
            int id = readInt("Введите ID желания для редактирования: ");
            Wish* wish = wishlist.findWish(id);
            if (!wish) { cout << "Желание не найдено.\n"; continue; }
            cout << "Оставьте поле пустым, чтобы не менять.\n";
            string newTitle = readString("Название (" + wish->title + "): ");
            string newDesc = readString("Описание (" + wish->description + "): ");
            string newCat = readString("Категория (" + wish->category + "): ");
            string newPriority = readString("Приоритет (1-3) сейчас: " + to_string(wish->priority) + ": ");
            string newPrice = readString("Цена (" + (wish->hasPrice ? to_string(wish->price) : "") + "): ");
            string newLink = readString("Ссылка (" + wish->link + "): ");
            string newFulfilled = readString("Статус (1-исполнено, 0-нет) сейчас: " + string(wish->fulfilled ? "1" : "0") + ": ");
            map<string, string> updates;
            if (!newTitle.empty()) updates["title"] = newTitle;
            if (!newDesc.empty()) updates["description"] = newDesc;
            if (!newCat.empty()) updates["category"] = newCat;
            if (!newPriority.empty()) updates["priority"] = newPriority;
            if (!newPrice.empty()) updates["price"] = newPrice;
            if (!newLink.empty()) updates["link"] = newLink;
            if (!newFulfilled.empty()) updates["fulfilled"] = newFulfilled;
            if (wishlist.editWish(id, updates)) {
                cout << "Желание обновлено.\n";
            } else {
                cout << "Ошибка обновления.\n";
            }
        } else if (choice == "8") {
            int id = readInt("Введите ID желания для удаления: ");
            if (wishlist.deleteWish(id)) {
                cout << "Желание удалено.\n";
            } else {
                cout << "Желание не найдено.\n";
            }
        } else if (choice == "9") {
            auto stats = wishlist.getStats();
            cout << "\n=== СТАТИСТИКА ===\n";
            cout << "Всего желаний: " << get<int>(stats["total"]) << "\n";
            cout << "Исполнено: " << get<int>(stats["fulfilled"]) << "\n";
            cout << "Не исполнено: " << get<int>(stats["unfulfilled"]) << "\n";
            cout << "Средняя цена: " << fixed << setprecision(2) << get<double>(stats["avg_price"]) << "\n";
            cout << "По категориям:\n";
            auto categories = get<map<string, int>>(stats["categories"]);
            for (const auto& [cat, count] : categories) cout << "  " << cat << ": " << count << "\n";
            cout << "По приоритетам:\n";
            auto priorities = get<map<int, int>>(stats["priorities"]);
            for (const auto& [p, count] : priorities) {
                string name;
                switch (p) { case 1: name = "Низкий"; break; case 2: name = "Средний"; break; case 3: name = "Высокий"; break; }
                cout << "  " << name << ": " << count << "\n";
            }
        } else if (choice == "10") {
            wishlist.saveToFile();
            cout << "Сохранено.\n";
        } else if (choice == "11") {
            wishlist.loadFromFile();
            cout << "Загружено.\n";
        } else if (choice == "12") {
            wishlist.exportCSV();
            cout << "Экспортировано в wishes_export.csv\n";
        } else if (choice == "13") {
            try {
                wishlist.importCSV();
                cout << "Импортировано из wishes_export.csv\n";
            } catch (const exception& e) {
                cout << "Ошибка импорта: " << e.what() << "\n";
            }
        } else {
            cout << "Неизвестная команда.\n";
        }
    }
    return 0;
}
