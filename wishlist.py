# wishlist.py
import json
import csv
from dataclasses import dataclass, asdict
from datetime import date
from typing import List, Optional
from pathlib import Path

@dataclass
class Wish:
    id: int
    title: str
    description: str
    category: str
    priority: int  # 1-3
    price: Optional[float]
    link: str
    fulfilled: bool
    added_date: str

class Wishlist:
    def __init__(self):
        self.wishes: List[Wish] = []
        self.next_id = 1

    def add_wish(self, title: str, description: str, category: str, priority: int,
                 price: Optional[float], link: str, fulfilled: bool = False) -> Wish:
        if not title or not category:
            raise ValueError("Название и категория не могут быть пустыми")
        if priority not in (1, 2, 3):
            raise ValueError("Приоритет должен быть 1, 2 или 3")
        if price is not None and price < 0:
            raise ValueError("Цена не может быть отрицательной")
        wish = Wish(
            id=self.next_id,
            title=title,
            description=description or "",
            category=category,
            priority=priority,
            price=price,
            link=link or "",
            fulfilled=fulfilled,
            added_date=date.today().isoformat()
        )
        self.wishes.append(wish)
        self.next_id += 1
        return wish

    def find_wish(self, wish_id: int) -> Optional[Wish]:
        return next((w for w in self.wishes if w.id == wish_id), None)

    def edit_wish(self, wish_id: int, **kwargs) -> bool:
        wish = self.find_wish(wish_id)
        if not wish:
            return False
        for key, value in kwargs.items():
            if hasattr(wish, key) and value is not None:
                setattr(wish, key, value)
        return True

    def delete_wish(self, wish_id: int) -> bool:
        wish = self.find_wish(wish_id)
        if wish:
            self.wishes.remove(wish)
            return True
        return False

    def toggle_fulfilled(self, wish_id: int) -> bool:
        wish = self.find_wish(wish_id)
        if not wish:
            return False
        wish.fulfilled = not wish.fulfilled
        return True

    def search_wishes(self, query: str) -> List[Wish]:
        q = query.lower()
        return [w for w in self.wishes if q in w.title.lower() or q in w.description.lower()]

    def filter_by_fulfilled(self, fulfilled: bool) -> List[Wish]:
        return [w for w in self.wishes if w.fulfilled == fulfilled]

    def filter_by_category(self, category: str) -> List[Wish]:
        return [w for w in self.wishes if w.category.lower() == category.lower()]

    def filter_by_priority(self, priority: int) -> List[Wish]:
        return [w for w in self.wishes if w.priority == priority]

    def sort_by_priority(self, reverse: bool = True) -> List[Wish]:
        return sorted(self.wishes, key=lambda w: w.priority, reverse=reverse)

    def sort_by_price(self, reverse: bool = False) -> List[Wish]:
        return sorted(self.wishes, key=lambda w: w.price if w.price is not None else 0, reverse=reverse)

    def get_stats(self) -> dict:
        total = len(self.wishes)
        fulfilled = len(self.filter_by_fulfilled(True))
        unfulfilled = total - fulfilled
        prices = [w.price for w in self.wishes if w.price is not None]
        avg_price = sum(prices) / len(prices) if prices else 0.0
        categories = {}
        priorities = {1: 0, 2: 0, 3: 0}
        for w in self.wishes:
            categories[w.category] = categories.get(w.category, 0) + 1
            priorities[w.priority] += 1
        return {
            "total": total,
            "fulfilled": fulfilled,
            "unfulfilled": unfulfilled,
            "avg_price": avg_price,
            "categories": categories,
            "priorities": priorities
        }

    def save_to_file(self, filename: str = "wishes_data.json") -> None:
        data = {"wishes": [asdict(w) for w in self.wishes]}
        with open(filename, "w", encoding="utf-8") as f:
            json.dump(data, f, ensure_ascii=False, indent=2)

    def load_from_file(self, filename: str = "wishes_data.json") -> None:
        path = Path(filename)
        if not path.exists():
            return
        with open(filename, "r", encoding="utf-8") as f:
            data = json.load(f)
            self.wishes.clear()
            for item in data.get("wishes", []):
                wish = Wish(
                    id=item["id"],
                    title=item["title"],
                    description=item.get("description", ""),
                    category=item["category"],
                    priority=item["priority"],
                    price=item.get("price"),
                    link=item.get("link", ""),
                    fulfilled=item["fulfilled"],
                    added_date=item["added_date"]
                )
                self.wishes.append(wish)
                if wish.id >= self.next_id:
                    self.next_id = wish.id + 1

    def export_csv(self, filename: str = "wishes_export.csv") -> None:
        with open(filename, "w", newline="", encoding="utf-8") as f:
            writer = csv.writer(f, delimiter=";")
            writer.writerow(["ID", "Название", "Описание", "Категория", "Приоритет", "Цена", "Ссылка", "Исполнено", "Дата добавления"])
            for w in self.wishes:
                writer.writerow([w.id, w.title, w.description, w.category, w.priority,
                                 w.price if w.price is not None else "", w.link,
                                 "Да" if w.fulfilled else "Нет", w.added_date])

    def import_csv(self, filename: str = "wishes_export.csv") -> None:
        path = Path(filename)
        if not path.exists():
            raise FileNotFoundError("Файл не найден")
        with open(filename, "r", encoding="utf-8") as f:
            reader = csv.DictReader(f, delimiter=";")
            for row in reader:
                try:
                    price = float(row["Цена"]) if row["Цена"] else None
                    self.add_wish(
                        title=row["Название"],
                        description=row["Описание"],
                        category=row["Категория"],
                        priority=int(row["Приоритет"]),
                        price=price,
                        link=row["Ссылка"],
                        fulfilled=row["Исполнено"] == "Да"
                    )
                except Exception as e:
                    print(f"Ошибка импорта строки: {e}")

def print_wish(wish: Wish) -> None:
    status = "✅ Исполнено" if wish.fulfilled else "⏳ Желаемое"
    priority_text = {1: "Низкий", 2: "Средний", 3: "Высокий"}[wish.priority]
    print(f"#{wish.id} - {wish.title} ({priority_text} приоритет)")
    if wish.description:
        print(f"   Описание: {wish.description}")
    print(f"   Категория: {wish.category}")
    if wish.price is not None:
        print(f"   Цена: {wish.price:.2f}")
    if wish.link:
        print(f"   Ссылка: {wish.link}")
    print(f"   {status}, Добавлен: {wish.added_date}")

def main():
    wishlist = Wishlist()
    wishlist.load_from_file()

    while True:
        print("\n===== ВИШЛИСТ (Python) =====")
        print("1. Добавить желание")
        print("2. Показать все желания")
        print("3. Показать неисполненные желания")
        print("4. Показать исполненные желания")
        print("5. Найти желания по названию")
        print("6. Отметить желание как исполненное")
        print("7. Редактировать желание")
        print("8. Удалить желание")
        print("9. Показать статистику")
        print("10. Сохранить в файл")
        print("11. Загрузить из файла")
        print("12. Экспорт в CSV")
        print("13. Импорт из CSV")
        print("0. Выход")
        choice = input("Выберите действие: ").strip()

        if choice == "0":
            break
        elif choice == "1":
            title = input("Название: ").strip()
            if not title:
                print("Название не может быть пустым.")
                continue
            description = input("Описание (необязательно): ").strip()
            category = input("Категория: ").strip()
            if not category:
                print("Категория не может быть пустой.")
                continue
            try:
                priority = int(input("Приоритет (1-низкий, 2-средний, 3-высокий): ").strip())
            except ValueError:
                priority = 2
            price_str = input("Цена (необязательно, число): ").strip()
            price = float(price_str) if price_str else None
            link = input("Ссылка (необязательно): ").strip()
            try:
                wish = wishlist.add_wish(title, description, category, priority, price, link)
                print(f"Желание добавлено с ID {wish.id}")
            except Exception as e:
                print("Ошибка:", e)
        elif choice == "2":
            if not wishlist.wishes:
                print("Нет желаний.")
            else:
                for w in wishlist.wishes:
                    print_wish(w)
        elif choice == "3":
            unfulfilled = wishlist.filter_by_fulfilled(False)
            if not unfulfilled:
                print("Нет неисполненных желаний.")
            else:
                for w in unfulfilled:
                    print_wish(w)
        elif choice == "4":
            fulfilled = wishlist.filter_by_fulfilled(True)
            if not fulfilled:
                print("Нет исполненных желаний.")
            else:
                for w in fulfilled:
                    print_wish(w)
        elif choice == "5":
            query = input("Введите часть названия или описания: ").strip()
            if not query:
                print("Введите текст.")
                continue
            results = wishlist.search_wishes(query)
            if not results:
                print("Желания не найдены.")
            else:
                for w in results:
                    print_wish(w)
        elif choice == "6":
            try:
                wid = int(input("Введите ID желания: ").strip())
            except ValueError:
                print("Некорректный ID.")
                continue
            if wishlist.toggle_fulfilled(wid):
                print("Статус желания изменён.")
            else:
                print("Желание не найдено.")
        elif choice == "7":
            try:
                wid = int(input("Введите ID желания для редактирования: ").strip())
            except ValueError:
                print("Некорректный ID.")
                continue
            wish = wishlist.find_wish(wid)
            if not wish:
                print("Желание не найдено.")
                continue
            print("Оставьте поле пустым, чтобы не менять.")
            new_title = input(f"Название ({wish.title}): ").strip()
            new_desc = input(f"Описание ({wish.description}): ").strip()
            new_cat = input(f"Категория ({wish.category}): ").strip()
            new_priority = input(f"Приоритет (1-3) сейчас: {wish.priority}: ").strip()
            new_price = input(f"Цена ({wish.price if wish.price is not None else ''}): ").strip()
            new_link = input(f"Ссылка ({wish.link}): ").strip()
            new_fulfilled = input(f"Статус (1-исполнено, 0-нет) сейчас: {'1' if wish.fulfilled else '0'}: ").strip()
            updates = {}
            if new_title: updates["title"] = new_title
            if new_desc: updates["description"] = new_desc
            if new_cat: updates["category"] = new_cat
            if new_priority:
                try:
                    updates["priority"] = int(new_priority)
                except ValueError:
                    print("Приоритет должен быть числом, пропускаем.")
            if new_price:
                try:
                    updates["price"] = float(new_price) if new_price else None
                except ValueError:
                    print("Цена должна быть числом, пропускаем.")
            if new_link: updates["link"] = new_link
            if new_fulfilled: updates["fulfilled"] = new_fulfilled == "1"
            if wishlist.edit_wish(wid, **updates):
                print("Желание обновлено.")
            else:
                print("Ошибка обновления.")
        elif choice == "8":
            try:
                wid = int(input("Введите ID желания для удаления: ").strip())
            except ValueError:
                print("Некорректный ID.")
                continue
            if wishlist.delete_wish(wid):
                print("Желание удалено.")
            else:
                print("Желание не найдено.")
        elif choice == "9":
            stats = wishlist.get_stats()
            print("\n=== СТАТИСТИКА ===")
            print(f"Всего желаний: {stats['total']}")
            print(f"Исполнено: {stats['fulfilled']}")
            print(f"Не исполнено: {stats['unfulfilled']}")
            print(f"Средняя цена: {stats['avg_price']:.2f}")
            print("По категориям:")
            for cat, count in stats['categories'].items():
                print(f"  {cat}: {count}")
            print("По приоритетам:")
            for p, count in stats['priorities'].items():
                name = {1: "Низкий", 2: "Средний", 3: "Высокий"}[p]
                print(f"  {name}: {count}")
        elif choice == "10":
            wishlist.save_to_file()
            print("Сохранено.")
        elif choice == "11":
            wishlist.load_from_file()
            print("Загружено.")
        elif choice == "12":
            wishlist.export_csv()
            print("Экспортировано в wishes_export.csv")
        elif choice == "13":
            try:
                wishlist.import_csv()
                print("Импортировано из wishes_export.csv")
            except FileNotFoundError:
                print("Файл wishes_export.csv не найден.")
            except Exception as e:
                print("Ошибка импорта:", e)
        else:
            print("Неизвестная команда.")

if __name__ == "__main__":
    main()
