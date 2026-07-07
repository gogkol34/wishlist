// wishlist.js
const fs = require('fs').promises;
const readline = require('readline');
const { parse } = require('csv-parse');
const { createObjectCsvWriter } = require('csv-writer');

const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
});

const question = (prompt) => new Promise(resolve => rl.question(prompt, resolve));

class Wish {
    constructor(id, title, description, category, priority, price, link, fulfilled, addedDate) {
        this.id = id;
        this.title = title;
        this.description = description || '';
        this.category = category;
        this.priority = priority;
        this.price = price;
        this.link = link || '';
        this.fulfilled = fulfilled || false;
        this.addedDate = addedDate || new Date().toISOString().slice(0, 10);
    }
}

class Wishlist {
    constructor() {
        this.wishes = [];
        this.nextId = 1;
    }

    addWish(title, description, category, priority, price, link, fulfilled = false) {
        if (!title.trim() || !category.trim()) throw new Error('Название и категория не могут быть пустыми');
        if (priority < 1 || priority > 3) throw new Error('Приоритет должен быть 1, 2 или 3');
        if (price !== null && price !== undefined && price < 0) throw new Error('Цена не может быть отрицательной');
        const wish = new Wish(this.nextId, title, description, category, priority, price, link, fulfilled);
        this.wishes.push(wish);
        this.nextId++;
        return wish;
    }

    findWish(id) {
        return this.wishes.find(w => w.id === id);
    }

    editWish(id, updates) {
        const wish = this.findWish(id);
        if (!wish) return false;
        Object.assign(wish, updates);
        return true;
    }

    deleteWish(id) {
        const index = this.wishes.findIndex(w => w.id === id);
        if (index === -1) return false;
        this.wishes.splice(index, 1);
        return true;
    }

    toggleFulfilled(id) {
        const wish = this.findWish(id);
        if (!wish) return false;
        wish.fulfilled = !wish.fulfilled;
        return true;
    }

    searchWishes(query) {
        const q = query.toLowerCase();
        return this.wishes.filter(w => w.title.toLowerCase().includes(q) || w.description.toLowerCase().includes(q));
    }

    filterByFulfilled(fulfilled) {
        return this.wishes.filter(w => w.fulfilled === fulfilled);
    }

    filterByCategory(category) {
        return this.wishes.filter(w => w.category.toLowerCase() === category.toLowerCase());
    }

    filterByPriority(priority) {
        return this.wishes.filter(w => w.priority === priority);
    }

    sortByPriority(reverse = true) {
        return [...this.wishes].sort((a, b) => reverse ? b.priority - a.priority : a.priority - b.priority);
    }

    sortByPrice(reverse = false) {
        return [...this.wishes].sort((a, b) => {
            const pa = a.price !== null ? a.price : 0;
            const pb = b.price !== null ? b.price : 0;
            return reverse ? pb - pa : pa - pb;
        });
    }

    getStats() {
        const total = this.wishes.length;
        const fulfilled = this.filterByFulfilled(true).length;
        const unfulfilled = total - fulfilled;
        const prices = this.wishes.filter(w => w.price !== null && w.price !== undefined).map(w => w.price);
        const avgPrice = prices.length ? prices.reduce((a, b) => a + b, 0) / prices.length : 0;
        const categories = {};
        const priorities = { 1: 0, 2: 0, 3: 0 };
        this.wishes.forEach(w => {
            categories[w.category] = (categories[w.category] || 0) + 1;
            priorities[w.priority]++;
        });
        return { total, fulfilled, unfulfilled, avgPrice, categories, priorities };
    }

    async saveToFile(filename = 'wishes_data.json') {
        const data = { wishes: this.wishes };
        await fs.writeFile(filename, JSON.stringify(data, null, 2));
    }

    async loadFromFile(filename = 'wishes_data.json') {
        try {
            const data = await fs.readFile(filename, 'utf8');
            const parsed = JSON.parse(data);
            this.wishes = parsed.wishes.map(w => Object.assign(new Wish(0), w));
            this.nextId = this.wishes.reduce((max, w) => Math.max(max, w.id), 0) + 1;
        } catch (err) {
            if (err.code !== 'ENOENT') throw err;
        }
    }

    async exportCSV(filename = 'wishes_export.csv') {
        const records = this.wishes.map(w => ({
            ID: w.id,
            Название: w.title,
            Описание: w.description,
            Категория: w.category,
            Приоритет: w.priority,
            Цена: w.price !== null ? w.price.toFixed(2) : '',
            Ссылка: w.link,
            Исполнено: w.fulfilled ? 'Да' : 'Нет',
            'Дата добавления': w.addedDate
        }));
        const csvWriter = createObjectCsvWriter({
            path: filename,
            header: [
                { id: 'ID', title: 'ID' },
                { id: 'Название', title: 'Название' },
                { id: 'Описание', title: 'Описание' },
                { id: 'Категория', title: 'Категория' },
                { id: 'Приоритет', title: 'Приоритет' },
                { id: 'Цена', title: 'Цена' },
                { id: 'Ссылка', title: 'Ссылка' },
                { id: 'Исполнено', title: 'Исполнено' },
                { id: 'Дата добавления', title: 'Дата добавления' }
            ],
            fieldDelimiter: ';'
        });
        await csvWriter.writeRecords(records);
    }

    async importCSV(filename = 'wishes_export.csv') {
        const fileContent = await fs.readFile(filename, 'utf8');
        return new Promise((resolve, reject) => {
            parse(fileContent, { columns: true, delimiter: ';' }, (err, records) => {
                if (err) reject(err);
                for (const row of records) {
                    try {
                        const price = row['Цена'] ? parseFloat(row['Цена']) : null;
                        this.addWish(
                            row['Название'],
                            row['Описание'],
                            row['Категория'],
                            parseInt(row['Приоритет']),
                            price,
                            row['Ссылка'],
                            row['Исполнено'] === 'Да'
                        );
                    } catch (e) {
                        console.log('Ошибка импорта строки:', e.message);
                    }
                }
                resolve();
            });
        });
    }
}

function printWish(wish) {
    const status = wish.fulfilled ? '✅ Исполнено' : '⏳ Желаемое';
    const priorityText = { 1: 'Низкий', 2: 'Средний', 3: 'Высокий' }[wish.priority];
    console.log(`#${wish.id} - ${wish.title} (${priorityText} приоритет)`);
    if (wish.description) console.log(`   Описание: ${wish.description}`);
    console.log(`   Категория: ${wish.category}`);
    if (wish.price !== null) console.log(`   Цена: ${wish.price.toFixed(2)}`);
    if (wish.link) console.log(`   Ссылка: ${wish.link}`);
    console.log(`   ${status}, Добавлен: ${wish.addedDate}`);
}

async function main() {
    const wishlist = new Wishlist();
    await wishlist.loadFromFile();

    while (true) {
        console.log('\n===== ВИШЛИСТ (JavaScript) =====');
        console.log('1. Добавить желание');
        console.log('2. Показать все желания');
        console.log('3. Показать неисполненные желания');
        console.log('4. Показать исполненные желания');
        console.log('5. Найти желания по названию');
        console.log('6. Отметить желание как исполненное');
        console.log('7. Редактировать желание');
        console.log('8. Удалить желание');
        console.log('9. Показать статистику');
        console.log('10. Сохранить в файл');
        console.log('11. Загрузить из файла');
        console.log('12. Экспорт в CSV');
        console.log('13. Импорт из CSV');
        console.log('0. Выход');
        const choice = await question('Выберите действие: ');

        if (choice === '0') break;

        switch (choice) {
            case '1': {
                const title = await question('Название: ');
                if (!title.trim()) { console.log('Название не может быть пустым.'); continue; }
                const description = await question('Описание (необязательно): ');
                const category = await question('Категория: ');
                if (!category.trim()) { console.log('Категория не может быть пустой.'); continue; }
                const priority = parseInt(await question('Приоритет (1-низкий, 2-средний, 3-высокий): '));
                const priceInput = await question('Цена (необязательно, число): ');
                const price = priceInput.trim() ? parseFloat(priceInput) : null;
                const link = await question('Ссылка (необязательно): ');
                try {
                    const wish = wishlist.addWish(title, description, category, priority, price, link);
                    console.log(`Желание добавлено с ID ${wish.id}`);
                } catch (err) {
                    console.log('Ошибка:', err.message);
                }
                break;
            }
            case '2':
                if (wishlist.wishes.length === 0) console.log('Нет желаний.');
                else wishlist.wishes.forEach(printWish);
                break;
            case '3': {
                const unfulfilled = wishlist.filterByFulfilled(false);
                if (unfulfilled.length === 0) console.log('Нет неисполненных желаний.');
                else unfulfilled.forEach(printWish);
                break;
            }
            case '4': {
                const fulfilled = wishlist.filterByFulfilled(true);
                if (fulfilled.length === 0) console.log('Нет исполненных желаний.');
                else fulfilled.forEach(printWish);
                break;
            }
            case '5': {
                const query = await question('Введите часть названия или описания: ');
                const results = wishlist.searchWishes(query);
                if (results.length === 0) console.log('Желания не найдены.');
                else results.forEach(printWish);
                break;
            }
            case '6': {
                const id = parseInt(await question('Введите ID желания: '));
                if (wishlist.toggleFulfilled(id)) {
                    console.log('Статус желания изменён.');
                } else {
                    console.log('Желание не найдено.');
                }
                break;
            }
            case '7': {
                const id = parseInt(await question('Введите ID желания для редактирования: '));
                const wish = wishlist.findWish(id);
                if (!wish) { console.log('Желание не найдено.'); continue; }
                console.log('Оставьте поле пустым, чтобы не менять.');
                const newTitle = await question(`Название (${wish.title}): `);
                const newDesc = await question(`Описание (${wish.description}): `);
                const newCat = await question(`Категория (${wish.category}): `);
                const newPriority = await question(`Приоритет (1-3) сейчас: ${wish.priority}: `);
                const newPrice = await question(`Цена (${wish.price !== null ? wish.price : ''}): `);
                const newLink = await question(`Ссылка (${wish.link}): `);
                const newFulfilled = await question(`Статус (1-исполнено, 0-нет) сейчас: ${wish.fulfilled ? '1' : '0'}: `);
                const updates = {};
                if (newTitle.trim()) updates.title = newTitle;
                if (newDesc.trim()) updates.description = newDesc;
                if (newCat.trim()) updates.category = newCat;
                if (newPriority.trim()) {
                    const p = parseInt(newPriority);
                    if (!isNaN(p)) updates.priority = p;
                    else console.log('Приоритет должен быть числом, пропускаем.');
                }
                if (newPrice.trim()) {
                    const p = parseFloat(newPrice);
                    if (!isNaN(p)) updates.price = p;
                    else console.log('Цена должна быть числом, пропускаем.');
                }
                if (newLink.trim()) updates.link = newLink;
                if (newFulfilled.trim()) updates.fulfilled = newFulfilled === '1';
                if (wishlist.editWish(id, updates)) console.log('Желание обновлено.');
                else console.log('Ошибка обновления.');
                break;
            }
            case '8': {
                const id = parseInt(await question('Введите ID желания для удаления: '));
                if (wishlist.deleteWish(id)) console.log('Желание удалено.');
                else console.log('Желание не найдено.');
                break;
            }
            case '9': {
                const stats = wishlist.getStats();
                console.log('\n=== СТАТИСТИКА ===');
                console.log(`Всего желаний: ${stats.total}`);
                console.log(`Исполнено: ${stats.fulfilled}`);
                console.log(`Не исполнено: ${stats.unfulfilled}`);
                console.log(`Средняя цена: ${stats.avgPrice.toFixed(2)}`);
                console.log('По категориям:');
                for (const [cat, count] of Object.entries(stats.categories)) {
                    console.log(`  ${cat}: ${count}`);
                }
                console.log('По приоритетам:');
                for (const [p, count] of Object.entries(stats.priorities)) {
                    const name = { 1: 'Низкий', 2: 'Средний', 3: 'Высокий' }[p];
                    console.log(`  ${name}: ${count}`);
                }
                break;
            }
            case '10':
                try {
                    await wishlist.saveToFile();
                    console.log('Сохранено.');
                } catch (err) {
                    console.log('Ошибка сохранения:', err.message);
                }
                break;
            case '11':
                try {
                    await wishlist.loadFromFile();
                    console.log('Загружено.');
                } catch (err) {
                    console.log('Ошибка загрузки:', err.message);
                }
                break;
            case '12':
                try {
                    await wishlist.exportCSV();
                    console.log('Экспортировано в wishes_export.csv');
                } catch (err) {
                    console.log('Ошибка экспорта:', err.message);
                }
                break;
            case '13':
                try {
                    await wishlist.importCSV();
                    console.log('Импортировано из wishes_export.csv');
                } catch (err) {
                    console.log('Ошибка импорта:', err.message);
                }
                break;
            default:
                console.log('Неизвестная команда.');
        }
    }
    rl.close();
}

main().catch(console.error);
