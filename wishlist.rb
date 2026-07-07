# wishlist.rb
require 'json'
require 'date'
require 'csv'

class Wish
  attr_accessor :id, :title, :description, :category, :priority, :price, :link, :fulfilled, :added_date

  def initialize(id, title, description, category, priority, price, link, fulfilled, added_date = Date.today.to_s)
    @id = id
    @title = title
    @description = description
    @category = category
    @priority = priority
    @price = price
    @link = link
    @fulfilled = fulfilled
    @added_date = added_date
  end

  def to_h
    { id: @id, title: @title, description: @description, category: @category,
      priority: @priority, price: @price, link: @link, fulfilled: @fulfilled, added_date: @added_date }
  end

  def self.from_h(hash)
    Wish.new(hash[:id], hash[:title], hash[:description], hash[:category], hash[:priority],
             hash[:price], hash[:link], hash[:fulfilled], hash[:added_date])
  end
end

class Wishlist
  attr_reader :wishes

  def initialize
    @wishes = []
    @next_id = 1
  end

  def add_wish(title, description, category, priority, price, link, fulfilled = false)
    raise "Название и категория не могут быть пустыми" if title.empty? || category.empty?
    raise "Приоритет должен быть 1, 2 или 3" unless [1, 2, 3].include?(priority)
    raise "Цена не может быть отрицательной" if price && price < 0
    wish = Wish.new(@next_id, title, description, category, priority, price, link, fulfilled)
    @wishes << wish
    @next_id += 1
    wish
  end

  def find_wish(id)
    @wishes.find { |w| w.id == id }
  end

  def edit_wish(id, **kwargs)
    wish = find_wish(id)
    return false unless wish
    kwargs.each do |key, value|
      wish.send("#{key}=", value) if wish.respond_to?("#{key}=")
    end
    true
  end

  def delete_wish(id)
    wish = find_wish(id)
    return false unless wish
    @wishes.delete(wish)
    true
  end

  def toggle_fulfilled(id)
    wish = find_wish(id)
    return false unless wish
    wish.fulfilled = !wish.fulfilled
    true
  end

  def search_wishes(query)
    q = query.downcase
    @wishes.select { |w| w.title.downcase.include?(q) || w.description.downcase.include?(q) }
  end

  def filter_by_fulfilled(fulfilled)
    @wishes.select { |w| w.fulfilled == fulfilled }
  end

  def filter_by_category(category)
    @wishes.select { |w| w.category.downcase == category.downcase }
  end

  def filter_by_priority(priority)
    @wishes.select { |w| w.priority == priority }
  end

  def sort_by_priority(reverse = true)
    @wishes.sort_by { |w| w.priority }.reverse! if reverse
    @wishes.sort_by { |w| w.priority }
  end

  def sort_by_price(reverse = false)
    @wishes.sort_by { |w| w.price ? w.price : 0 }
    @wishes.reverse! if reverse
    @wishes
  end

  def stats
    total = @wishes.size
    fulfilled = filter_by_fulfilled(true).size
    unfulfilled = total - fulfilled
    prices = @wishes.map(&:price).compact
    avg_price = prices.empty? ? 0 : prices.sum.to_f / prices.size
    categories = Hash.new(0)
    priorities = { 1 => 0, 2 => 0, 3 => 0 }
    @wishes.each do |w|
      categories[w.category] += 1
      priorities[w.priority] += 1
    end
    { total: total, fulfilled: fulfilled, unfulfilled: unfulfilled, avg_price: avg_price,
      categories: categories, priorities: priorities }
  end

  def save_to_file(filename = "wishes_data.json")
    data = { wishes: @wishes.map(&:to_h) }
    File.write(filename, JSON.pretty_generate(data))
  end

  def load_from_file(filename = "wishes_data.json")
    return unless File.exist?(filename)
    data = JSON.parse(File.read(filename), symbolize_names: true)
    @wishes.clear
    data[:wishes].each do |item|
      wish = Wish.from_h(item)
      @wishes << wish
      @next_id = wish.id + 1 if wish.id >= @next_id
    end
  rescue JSON::ParserError
    puts "Ошибка чтения файла."
  end

  def export_csv(filename = "wishes_export.csv")
    CSV.open(filename, "w", col_sep: ";") do |csv|
      csv << ["ID", "Название", "Описание", "Категория", "Приоритет", "Цена", "Ссылка", "Исполнено", "Дата добавления"]
      @wishes.each do |w|
        csv << [w.id, w.title, w.description, w.category, w.priority,
                w.price ? w.price : "", w.link, w.fulfilled ? "Да" : "Нет", w.added_date]
      end
    end
  end

  def import_csv(filename = "wishes_export.csv")
    unless File.exist?(filename)
      raise "Файл не найден"
    end
    CSV.foreach(filename, headers: true, col_sep: ";") do |row|
      begin
        price = row["Цена"].empty? ? nil : row["Цена"].to_f
        add_wish(
          title: row["Название"],
          description: row["Описание"],
          category: row["Категория"],
          priority: row["Приоритет"].to_i,
          price: price,
          link: row["Ссылка"],
          fulfilled: row["Исполнено"] == "Да"
        )
      rescue => e
        puts "Ошибка импорта строки: #{e}"
      end
    end
  end
end

def print_wish(wish)
  status = wish.fulfilled ? "✅ Исполнено" : "⏳ Желаемое"
  priority_text = { 1 => "Низкий", 2 => "Средний", 3 => "Высокий" }[wish.priority]
  puts "##{wish.id} - #{wish.title} (#{priority_text} приоритет)"
  puts "   Описание: #{wish.description}" unless wish.description.empty?
  puts "   Категория: #{wish.category}"
  puts "   Цена: #{'%.2f' % wish.price}" if wish.price
  puts "   Ссылка: #{wish.link}" unless wish.link.empty?
  puts "   #{status}, Добавлен: #{wish.added_date}"
end

def main
  wishlist = Wishlist.new
  wishlist.load_from_file

  loop do
    puts "\n===== ВИШЛИСТ (Ruby) ====="
    puts "1. Добавить желание"
    puts "2. Показать все желания"
    puts "3. Показать неисполненные желания"
    puts "4. Показать исполненные желания"
    puts "5. Найти желания по названию"
    puts "6. Отметить желание как исполненное"
    puts "7. Редактировать желание"
    puts "8. Удалить желание"
    puts "9. Показать статистику"
    puts "10. Сохранить в файл"
    puts "11. Загрузить из файла"
    puts "12. Экспорт в CSV"
    puts "13. Импорт из CSV"
    puts "0. Выход"
    print "Выберите действие: "
    choice = gets.chomp

    case choice
    when "0"
      break
    when "1"
      print "Название: "
      title = gets.chomp
      next if title.empty?
      print "Описание (необязательно): "
      description = gets.chomp
      print "Категория: "
      category = gets.chomp
      next if category.empty?
      print "Приоритет (1-низкий, 2-средний, 3-высокий): "
      priority = gets.chomp.to_i
      print "Цена (необязательно, число): "
      price_input = gets.chomp
      price = price_input.empty? ? nil : price_input.to_f
      print "Ссылка (необязательно): "
      link = gets.chomp
      begin
        wish = wishlist.add_wish(title, description, category, priority, price, link)
        puts "Желание добавлено с ID #{wish.id}"
      rescue => e
        puts "Ошибка: #{e.message}"
      end
    when "2"
      if wishlist.wishes.empty?
        puts "Нет желаний."
      else
        wishlist.wishes.each { |w| print_wish(w) }
      end
    when "3"
      unfulfilled = wishlist.filter_by_fulfilled(false)
      if unfulfilled.empty?
        puts "Нет неисполненных желаний."
      else
        unfulfilled.each { |w| print_wish(w) }
      end
    when "4"
      fulfilled = wishlist.filter_by_fulfilled(true)
      if fulfilled.empty?
        puts "Нет исполненных желаний."
      else
        fulfilled.each { |w| print_wish(w) }
      end
    when "5"
      print "Введите часть названия или описания: "
      query = gets.chomp
      results = wishlist.search_wishes(query)
      if results.empty?
        puts "Желания не найдены."
      else
        results.each { |w| print_wish(w) }
      end
    when "6"
      print "Введите ID желания: "
      id = gets.chomp.to_i
      if wishlist.toggle_fulfilled(id)
        puts "Статус желания изменён."
      else
        puts "Желание не найдено."
      end
    when "7"
      print "Введите ID желания для редактирования: "
      id = gets.chomp.to_i
      wish = wishlist.find_wish(id)
      unless wish
        puts "Желание не найдено."
        next
      end
      puts "Оставьте поле пустым, чтобы не менять."
      print "Название (#{wish.title}): "
      new_title = gets.chomp
      print "Описание (#{wish.description}): "
      new_desc = gets.chomp
      print "Категория (#{wish.category}): "
      new_cat = gets.chomp
      print "Приоритет (1-3) сейчас: #{wish.priority}: "
      new_priority = gets.chomp
      print "Цена (#{wish.price ? wish.price : ''}): "
      new_price = gets.chomp
      print "Ссылка (#{wish.link}): "
      new_link = gets.chomp
      print "Статус (1-исполнено, 0-нет) сейчас: #{wish.fulfilled ? '1' : '0'}: "
      new_fulfilled = gets.chomp
      updates = {}
      updates[:title] = new_title unless new_title.empty?
      updates[:description] = new_desc unless new_desc.empty?
      updates[:category] = new_cat unless new_cat.empty?
      unless new_priority.empty?
        updates[:priority] = new_priority.to_i
      end
      unless new_price.empty?
        updates[:price] = new_price.empty? ? nil : new_price.to_f
      end
      updates[:link] = new_link unless new_link.empty?
      unless new_fulfilled.empty?
        updates[:fulfilled] = new_fulfilled == "1"
      end
      if wishlist.edit_wish(id, **updates)
        puts "Желание обновлено."
      else
        puts "Ошибка обновления."
      end
    when "8"
      print "Введите ID желания для удаления: "
      id = gets.chomp.to_i
      if wishlist.delete_wish(id)
        puts "Желание удалено."
      else
        puts "Желание не найдено."
      end
    when "9"
      stats = wishlist.stats
      puts "\n=== СТАТИСТИКА ==="
      puts "Всего желаний: #{stats[:total]}"
      puts "Исполнено: #{stats[:fulfilled]}"
      puts "Не исполнено: #{stats[:unfulfilled]}"
      puts "Средняя цена: #{'%.2f' % stats[:avg_price]}"
      puts "По категориям:"
      stats[:categories].each { |cat, count| puts "  #{cat}: #{count}" }
      puts "По приоритетам:"
      { 1 => "Низкий", 2 => "Средний", 3 => "Высокий" }.each do |p, name|
        puts "  #{name}: #{stats[:priorities][p]}"
      end
    when "10"
      wishlist.save_to_file
      puts "Сохранено."
    when "11"
      wishlist.load_from_file
      puts "Загружено."
    when "12"
      wishlist.export_csv
      puts "Экспортировано в wishes_export.csv"
    when "13"
      begin
        wishlist.import_csv
        puts "Импортировано из wishes_export.csv"
      rescue => e
        puts "Ошибка импорта: #{e}"
      end
    else
      puts "Неизвестная команда."
    end
  end
end

main if __FILE__ == $0
