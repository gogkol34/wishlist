// wishlist.go
package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Wish struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Priority    int     `json:"priority"`
	Price       *float64 `json:"price,omitempty"`
	Link        string  `json:"link"`
	Fulfilled   bool    `json:"fulfilled"`
	AddedDate   string  `json:"added_date"`
}

type WishesData struct {
	Wishes []Wish `json:"wishes"`
}

type Wishlist struct {
	wishes  []Wish
	nextID  int
}

func NewWishlist() *Wishlist {
	return &Wishlist{
		wishes:  []Wish{},
		nextID:  1,
	}
}

func (wl *Wishlist) AddWish(title, description, category string, priority int, price *float64, link string, fulfilled bool) (Wish, error) {
	if title == "" || category == "" {
		return Wish{}, fmt.Errorf("название и категория не могут быть пустыми")
	}
	if priority < 1 || priority > 3 {
		return Wish{}, fmt.Errorf("приоритет должен быть 1, 2 или 3")
	}
	if price != nil && *price < 0 {
		return Wish{}, fmt.Errorf("цена не может быть отрицательной")
	}
	wish := Wish{
		ID:          wl.nextID,
		Title:       title,
		Description: description,
		Category:    category,
		Priority:    priority,
		Price:       price,
		Link:        link,
		Fulfilled:   fulfilled,
		AddedDate:   time.Now().Format("2006-01-02"),
	}
	wl.wishes = append(wl.wishes, wish)
	wl.nextID++
	return wish, nil
}

func (wl *Wishlist) FindWish(id int) *Wish {
	for i := range wl.wishes {
		if wl.wishes[i].ID == id {
			return &wl.wishes[i]
		}
	}
	return nil
}

func (wl *Wishlist) EditWish(id int, updates map[string]interface{}) bool {
	wish := wl.FindWish(id)
	if wish == nil {
		return false
	}
	for key, value := range updates {
		switch key {
		case "title":
			if v, ok := value.(string); ok {
				wish.Title = v
			}
		case "description":
			if v, ok := value.(string); ok {
				wish.Description = v
			}
		case "category":
			if v, ok := value.(string); ok {
				wish.Category = v
			}
		case "priority":
			if v, ok := value.(int); ok {
				wish.Priority = v
			}
		case "price":
			if v, ok := value.(*float64); ok {
				wish.Price = v
			}
		case "link":
			if v, ok := value.(string); ok {
				wish.Link = v
			}
		case "fulfilled":
			if v, ok := value.(bool); ok {
				wish.Fulfilled = v
			}
		}
	}
	return true
}

func (wl *Wishlist) DeleteWish(id int) bool {
	for i, w := range wl.wishes {
		if w.ID == id {
			wl.wishes = append(wl.wishes[:i], wl.wishes[i+1:]...)
			return true
		}
	}
	return false
}

func (wl *Wishlist) ToggleFulfilled(id int) bool {
	wish := wl.FindWish(id)
	if wish == nil {
		return false
	}
	wish.Fulfilled = !wish.Fulfilled
	return true
}

func (wl *Wishlist) SearchWishes(query string) []Wish {
	q := strings.ToLower(query)
	var result []Wish
	for _, w := range wl.wishes {
		if strings.Contains(strings.ToLower(w.Title), q) || strings.Contains(strings.ToLower(w.Description), q) {
			result = append(result, w)
		}
	}
	return result
}

func (wl *Wishlist) FilterByFulfilled(fulfilled bool) []Wish {
	var result []Wish
	for _, w := range wl.wishes {
		if w.Fulfilled == fulfilled {
			result = append(result, w)
		}
	}
	return result
}

func (wl *Wishlist) FilterByCategory(category string) []Wish {
	var result []Wish
	for _, w := range wl.wishes {
		if strings.EqualFold(w.Category, category) {
			result = append(result, w)
		}
	}
	return result
}

func (wl *Wishlist) FilterByPriority(priority int) []Wish {
	var result []Wish
	for _, w := range wl.wishes {
		if w.Priority == priority {
			result = append(result, w)
		}
	}
	return result
}

func (wl *Wishlist) SortByPriority(reverse bool) []Wish {
	sorted := make([]Wish, len(wl.wishes))
	copy(sorted, wl.wishes)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if reverse {
				if sorted[i].Priority < sorted[j].Priority {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			} else {
				if sorted[i].Priority > sorted[j].Priority {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
	}
	return sorted
}

func (wl *Wishlist) SortByPrice(reverse bool) []Wish {
	sorted := make([]Wish, len(wl.wishes))
	copy(sorted, wl.wishes)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			priceI := 0.0
			if sorted[i].Price != nil {
				priceI = *sorted[i].Price
			}
			priceJ := 0.0
			if sorted[j].Price != nil {
				priceJ = *sorted[j].Price
			}
			if reverse {
				if priceI < priceJ {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			} else {
				if priceI > priceJ {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
	}
	return sorted
}

func (wl *Wishlist) GetStats() map[string]interface{} {
	total := len(wl.wishes)
	fulfilled := len(wl.FilterByFulfilled(true))
	unfulfilled := total - fulfilled
	var prices []float64
	for _, w := range wl.wishes {
		if w.Price != nil {
			prices = append(prices, *w.Price)
		}
	}
	avgPrice := 0.0
	if len(prices) > 0 {
		sum := 0.0
		for _, p := range prices {
			sum += p
		}
		avgPrice = sum / float64(len(prices))
	}
	categories := make(map[string]int)
	priorities := map[int]int{1: 0, 2: 0, 3: 0}
	for _, w := range wl.wishes {
		categories[w.Category]++
		priorities[w.Priority]++
	}
	return map[string]interface{}{
		"total":       total,
		"fulfilled":   fulfilled,
		"unfulfilled": unfulfilled,
		"avg_price":   avgPrice,
		"categories":  categories,
		"priorities":  priorities,
	}
}

func (wl *Wishlist) SaveToFile(filename string) error {
	data := WishesData{Wishes: wl.wishes}
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}

func (wl *Wishlist) LoadFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var wd WishesData
	if err := json.Unmarshal(data, &wd); err != nil {
		return err
	}
	wl.wishes = wd.Wishes
	for _, w := range wl.wishes {
		if w.ID >= wl.nextID {
			wl.nextID = w.ID + 1
		}
	}
	return nil
}

func (wl *Wishlist) ExportCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	writer.Comma = ';'
	defer writer.Flush()
	headers := []string{"ID", "Название", "Описание", "Категория", "Приоритет", "Цена", "Ссылка", "Исполнено", "Дата добавления"}
	if err := writer.Write(headers); err != nil {
		return err
	}
	for _, w := range wl.wishes {
		priceStr := ""
		if w.Price != nil {
			priceStr = fmt.Sprintf("%.2f", *w.Price)
		}
		fulfilledStr := "Нет"
		if w.Fulfilled {
			fulfilledStr = "Да"
		}
		row := []string{
			strconv.Itoa(w.ID),
			w.Title,
			w.Description,
			w.Category,
			strconv.Itoa(w.Priority),
			priceStr,
			w.Link,
			fulfilledStr,
			w.AddedDate,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func (wl *Wishlist) ImportCSV(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.LazyQuotes = true
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return fmt.Errorf("файл пуст или нет данных")
	}
	for _, row := range records[1:] {
		if len(row) < 9 {
			continue
		}
		title := row[1]
		description := row[2]
		category := row[3]
		priority, _ := strconv.Atoi(row[4])
		var price *float64
		if row[5] != "" {
			p, _ := strconv.ParseFloat(row[5], 64)
			price = &p
		}
		link := row[6]
		fulfilled := row[7] == "Да"
		_, err := wl.AddWish(title, description, category, priority, price, link, fulfilled)
		if err != nil {
			fmt.Println("Ошибка импорта строки:", err)
		}
	}
	return nil
}

func readString(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func readInt(prompt string) int {
	for {
		input := readString(prompt)
		if val, err := strconv.Atoi(input); err == nil {
			return val
		}
		fmt.Println("Введите число.")
	}
}

func readFloat(prompt string) *float64 {
	input := readString(prompt)
	if input == "" {
		return nil
	}
	val, err := strconv.ParseFloat(input, 64)
	if err != nil {
		fmt.Println("Ошибка ввода, пропускаем.")
		return nil
	}
	return &val
}

func readBool(prompt string) bool {
	for {
		input := readString(prompt)
		if input == "1" {
			return true
		} else if input == "0" {
			return false
		}
		fmt.Println("Введите 1 или 0.")
	}
}

func printWish(wish Wish) {
	status := "⏳ Желаемое"
	if wish.Fulfilled {
		status = "✅ Исполнено"
	}
	priorityText := map[int]string{1: "Низкий", 2: "Средний", 3: "Высокий"}[wish.Priority]
	fmt.Printf("#%d - %s (%s приоритет)\n", wish.ID, wish.Title, priorityText)
	if wish.Description != "" {
		fmt.Printf("   Описание: %s\n", wish.Description)
	}
	fmt.Printf("   Категория: %s\n", wish.Category)
	if wish.Price != nil {
		fmt.Printf("   Цена: %.2f\n", *wish.Price)
	}
	if wish.Link != "" {
		fmt.Printf("   Ссылка: %s\n", wish.Link)
	}
	fmt.Printf("   %s, Добавлен: %s\n", status, wish.AddedDate)
}

func main() {
	wishlist := NewWishlist()
	if err := wishlist.LoadFromFile("wishes_data.json"); err != nil {
		fmt.Println("Ошибка загрузки:", err)
	}

	for {
		fmt.Println("\n===== ВИШЛИСТ (Go) =====")
		fmt.Println("1. Добавить желание")
		fmt.Println("2. Показать все желания")
		fmt.Println("3. Показать неисполненные желания")
		fmt.Println("4. Показать исполненные желания")
		fmt.Println("5. Найти желания по названию")
		fmt.Println("6. Отметить желание как исполненное")
		fmt.Println("7. Редактировать желание")
		fmt.Println("8. Удалить желание")
		fmt.Println("9. Показать статистику")
		fmt.Println("10. Сохранить в файл")
		fmt.Println("11. Загрузить из файла")
		fmt.Println("12. Экспорт в CSV")
		fmt.Println("13. Импорт из CSV")
		fmt.Println("0. Выход")
		choice := readString("Выберите действие: ")

		switch choice {
		case "0":
			return
		case "1":
			title := readString("Название: ")
			if title == "" {
				fmt.Println("Название не может быть пустым.")
				continue
			}
			description := readString("Описание (необязательно): ")
			category := readString("Категория: ")
			if category == "" {
				fmt.Println("Категория не может быть пустой.")
				continue
			}
			priority := readInt("Приоритет (1-низкий, 2-средний, 3-высокий): ")
			price := readFloat("Цена (необязательно, число): ")
			link := readString("Ссылка (необязательно): ")
			wish, err := wishlist.AddWish(title, description, category, priority, price, link, false)
			if err != nil {
				fmt.Println("Ошибка:", err)
			} else {
				fmt.Printf("Желание добавлено с ID %d\n", wish.ID)
			}
		case "2":
			if len(wishlist.wishes) == 0 {
				fmt.Println("Нет желаний.")
			} else {
				for _, w := range wishlist.wishes {
					printWish(w)
				}
			}
		case "3":
			unfulfilled := wishlist.FilterByFulfilled(false)
			if len(unfulfilled) == 0 {
				fmt.Println("Нет неисполненных желаний.")
			} else {
				for _, w := range unfulfilled {
					printWish(w)
				}
			}
		case "4":
			fulfilled := wishlist.FilterByFulfilled(true)
			if len(fulfilled) == 0 {
				fmt.Println("Нет исполненных желаний.")
			} else {
				for _, w := range fulfilled {
					printWish(w)
				}
			}
		case "5":
			query := readString("Введите часть названия или описания: ")
			results := wishlist.SearchWishes(query)
			if len(results) == 0 {
				fmt.Println("Желания не найдены.")
			} else {
				for _, w := range results {
					printWish(w)
				}
			}
		case "6":
			id := readInt("Введите ID желания: ")
			if wishlist.ToggleFulfilled(id) {
				fmt.Println("Статус желания изменён.")
			} else {
				fmt.Println("Желание не найдено.")
			}
		case "7":
			id := readInt("Введите ID желания для редактирования: ")
			wish := wishlist.FindWish(id)
			if wish == nil {
				fmt.Println("Желание не найдено.")
				continue
			}
			fmt.Println("Оставьте поле пустым, чтобы не менять.")
			newTitle := readString(fmt.Sprintf("Название (%s): ", wish.Title))
			newDesc := readString(fmt.Sprintf("Описание (%s): ", wish.Description))
			newCat := readString(fmt.Sprintf("Категория (%s): ", wish.Category))
			newPriority := readString(fmt.Sprintf("Приоритет (1-3) сейчас: %d: ", wish.Priority))
			newPrice := readString(fmt.Sprintf("Цена (%v): ", func() string {
				if wish.Price != nil {
					return fmt.Sprintf("%.2f", *wish.Price)
				}
				return ""
			}()))
			newLink := readString(fmt.Sprintf("Ссылка (%s): ", wish.Link))
			newFulfilled := readString(fmt.Sprintf("Статус (1-исполнено, 0-нет) сейчас: %d: ", map[bool]int{true: 1, false: 0}[wish.Fulfilled]))
			updates := make(map[string]interface{})
			if newTitle != "" {
				updates["title"] = newTitle
			}
			if newDesc != "" {
				updates["description"] = newDesc
			}
			if newCat != "" {
				updates["category"] = newCat
			}
			if newPriority != "" {
				if p, err := strconv.Atoi(newPriority); err == nil {
					updates["priority"] = p
				} else {
					fmt.Println("Приоритет должен быть числом, пропускаем.")
				}
			}
			if newPrice != "" {
				if p, err := strconv.ParseFloat(newPrice, 64); err == nil {
					updates["price"] = &p
				} else {
					fmt.Println("Цена должна быть числом, пропускаем.")
				}
			}
			if newLink != "" {
				updates["link"] = newLink
			}
			if newFulfilled != "" {
				updates["fulfilled"] = newFulfilled == "1"
			}
			if wishlist.EditWish(id, updates) {
				fmt.Println("Желание обновлено.")
			} else {
				fmt.Println("Ошибка обновления.")
			}
		case "8":
			id := readInt("Введите ID желания для удаления: ")
			if wishlist.DeleteWish(id) {
				fmt.Println("Желание удалено.")
			} else {
				fmt.Println("Желание не найдено.")
			}
		case "9":
			stats := wishlist.GetStats()
			fmt.Println("\n=== СТАТИСТИКА ===")
			fmt.Printf("Всего желаний: %d\n", stats["total"])
			fmt.Printf("Исполнено: %d\n", stats["fulfilled"])
			fmt.Printf("Не исполнено: %d\n", stats["unfulfilled"])
			fmt.Printf("Средняя цена: %.2f\n", stats["avg_price"])
			fmt.Println("По категориям:")
			categories := stats["categories"].(map[string]int)
			for cat, count := range categories {
				fmt.Printf("  %s: %d\n", cat, count)
			}
			fmt.Println("По приоритетам:")
			priorities := stats["priorities"].(map[int]int)
			for p, count := range priorities {
				name := map[int]string{1: "Низкий", 2: "Средний", 3: "Высокий"}[p]
				fmt.Printf("  %s: %d\n", name, count)
			}
		case "10":
			if err := wishlist.SaveToFile("wishes_data.json"); err != nil {
				fmt.Println("Ошибка сохранения:", err)
			} else {
				fmt.Println("Сохранено.")
			}
		case "11":
			if err := wishlist.LoadFromFile("wishes_data.json"); err != nil {
				fmt.Println("Ошибка загрузки:", err)
			} else {
				fmt.Println("Загружено.")
			}
		case "12":
			if err := wishlist.ExportCSV("wishes_export.csv"); err != nil {
				fmt.Println("Ошибка экспорта:", err)
			} else {
				fmt.Println("Экспортировано в wishes_export.csv")
			}
		case "13":
			if err := wishlist.ImportCSV("wishes_export.csv"); err != nil {
				fmt.Println("Ошибка импорта:", err)
			} else {
				fmt.Println("Импортировано из wishes_export.csv")
			}
		default:
			fmt.Println("Неизвестная команда.")
		}
	}
}
