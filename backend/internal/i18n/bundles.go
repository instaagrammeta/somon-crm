package i18n

// Translation bundles. Keys are stable, locale-agnostic identifiers.
var bundles = map[string]map[string]string{
	LocaleTG: {
		// Generic
		"ok":                 "Бомуваффақият",
		"error.unknown":      "Хатогии номаълум",
		"error.bad_request":  "Дархости нодуруст",
		"error.unauthorized": "Сертификати воридшавӣ лозим аст",
		"error.forbidden":    "Дастрасӣ манъ аст",
		"error.not_found":    "Ёфт нашуд",
		"error.conflict":     "Ин маълумот аллакай вуҷуд дорад",
		"error.validation":   "Маълумоти ворид нодуруст: {0}",
		"error.internal":     "Хатогии дохилии сервер",

		// Auth
		"auth.invalid_credentials": "Логин ё парол нодуруст аст",
		"auth.token_expired":       "Сертификати воридшавӣ ба охир расид",
		"auth.token_invalid":       "Сертификати воридшавӣ нодуруст аст",
		"auth.logged_out":          "Шумо аз система баромадед",

		// Telegram
		"tg.linked":           "✅ Шумо ба система пайваст шудед: {0}",
		"tg.start_prompt":     "Салом! Барои пайваст кардани аккаунти CRM, рамзи пайвасткуниро аз администратор гиред ва ин фармонро равон кунед: /link CODE",
		"tg.link_invalid":     "❌ Рамз нодуруст аст ё муҳлаташ гузашт",
		"tg.task_assigned":    "📋 Ба шумо вазифаи нав таъин шуд:\n\n*{0}*\n{1}",
		"tg.task_updated":     "✏️ Вазифаи шумо навсозӣ шуд:\n\n*{0}*\nҲолат: {1}",
		"tg.request_assigned": "📥 Ба шумо заявкаи нав таъин шуд:\n\n👤 Мизоҷ: {0}\n📞 Телефон: {1}\n🏠 Суроға: {2}",
		"tg.lead_moved":       "🔄 Лиди шумо ба колонкаи нав ҳаракат кард: *{0}* → *{1}*",
		"tg.tariff_expiring":  "⚠️ Тарифи симкортаи {0} баъд аз {1} рӯз ба охир мерасад",

		// Entities
		"user.not_found":        "Корбар ёфт нашуд",
		"user.exists":           "Корбар бо чунин логин аллакай ҳаст",
		"task.not_found":        "Вазифа ёфт нашуд",
		"request.not_found":     "Заявка ёфт нашуд",
		"board.not_found":       "Доска ёфт нашуд",
		"column.not_found":      "Колонка ёфт нашуд",
		"lead.not_found":        "Лид ёфт нашуд",
		"house.not_found":       "Объекти манзил ёфт нашуд",
		"realty.not_found":      "Объекти бино ёфт нашуд",
		"objekt.not_found":      "Объекти шахматка ёфт нашуд",
		"sim.not_found":         "Симкорт ёфт нашуд",
		"phone.not_found":       "Телефон ёфт нашуд",
		"tariff.not_found":      "Тариф ёфт нашуд",
		"post.not_found":        "Пост ёфт нашуд",
		"folder.not_found":      "Папка ёфт нашуд",
		"file.not_found":        "Файл ёфт нашуд",
		"bank.not_found":        "Бонк ёфт нашуд",
		"installment.not_found": "Объекти рассрочка ёфт нашуд",

		// In-app notifications
		"notify.task_assigned":       "Ба шумо вазифаи нав таъин шуд",
		"notify.task_status_changed": "Ҳолати вазифа тағйир ёфт",
		"notify.request_assigned":    "Ба шумо заявкаи нав таъин шуд",
		"notify.lead_moved":          "Лиди шумо ҳаракат кард",
		"notify.tariff_expiring":     "Мӯҳлати тарифи симкорт ба охир мерасад",

		// Task status labels (used inside notification text)
		"task.status_new":         "Нав",
		"task.status_in_progress": "Дар ҷараён",
		"task.status_done":        "Анҷомёфта",
		"task.status_cancelled":   "Бекоршуда",
	},
	LocaleRU: {
		"ok":                 "Успешно",
		"error.unknown":      "Неизвестная ошибка",
		"error.bad_request":  "Неверный запрос",
		"error.unauthorized": "Требуется авторизация",
		"error.forbidden":    "Доступ запрещён",
		"error.not_found":    "Не найдено",
		"error.conflict":     "Запись уже существует",
		"error.validation":   "Ошибка валидации: {0}",
		"error.internal":     "Внутренняя ошибка сервера",

		"auth.invalid_credentials": "Неверный логин или пароль",
		"auth.token_expired":       "Срок действия токена истёк",
		"auth.token_invalid":       "Недействительный токен",
		"auth.logged_out":          "Вы вышли из системы",

		"tg.linked":           "✅ Вы успешно подключили аккаунт: {0}",
		"tg.start_prompt":     "Привет! Чтобы привязать аккаунт CRM, получите код у администратора и отправьте: /link CODE",
		"tg.link_invalid":     "❌ Неверный или просроченный код",
		"tg.task_assigned":    "📋 Вам назначена новая задача:\n\n*{0}*\n{1}",
		"tg.task_updated":     "✏️ Задача обновлена:\n\n*{0}*\nСтатус: {1}",
		"tg.request_assigned": "📥 Вам назначена новая заявка:\n\n👤 Клиент: {0}\n📞 Телефон: {1}\n🏠 Адрес: {2}",
		"tg.lead_moved":       "🔄 Ваш лид перемещён: *{0}* → *{1}*",
		"tg.tariff_expiring":  "⚠️ Тариф SIM-карты {0} истекает через {1} дней",

		"user.not_found":        "Пользователь не найден",
		"user.exists":           "Пользователь с таким логином уже существует",
		"task.not_found":        "Задача не найдена",
		"request.not_found":     "Заявка не найдена",
		"board.not_found":       "Доска не найдена",
		"column.not_found":      "Колонка не найдена",
		"lead.not_found":        "Лид не найден",
		"house.not_found":       "Объект недвижимости не найден",
		"realty.not_found":      "Объект не найден",
		"objekt.not_found":      "Проект шахматки не найден",
		"sim.not_found":         "SIM-карта не найдена",
		"phone.not_found":       "Телефон не найден",
		"tariff.not_found":      "Тариф не найден",
		"post.not_found":        "Пост не найден",
		"folder.not_found":      "Папка не найдена",
		"file.not_found":        "Файл не найден",
		"bank.not_found":        "Банк не найден",
		"installment.not_found": "Объект рассрочки не найден",

		// In-app notifications
		"notify.task_assigned":       "Вам назначена новая задача",
		"notify.task_status_changed": "Статус задачи изменён",
		"notify.request_assigned":    "Вам назначена новая заявка",
		"notify.lead_moved":          "Ваш лид перемещён",
		"notify.tariff_expiring":     "Срок тарифа SIM-карты истекает",

		// Task status labels (used inside notification text)
		"task.status_new":         "Новая",
		"task.status_in_progress": "В процессе",
		"task.status_done":        "Выполнено",
		"task.status_cancelled":   "Отменено",
	},
}
