package i18n

var messages = map[string]map[string]string{
	"ru": {
		"choose_lang":       "Пожалуйста, выберите язык обслуживания:",
		"ask_name":          "Здравствуйте! Пожалуйста, напишите ваше Имя и Фамилию.\nПример: Иван Иванов",
		"btn_share_contact": "📱 Поделиться контактом",
		"ask_phone":         "Отлично! Теперь отправьте или напишите ваш номер телефона.\nПример: +998901234567",
		"reg_success":       "Регистрация успешно завершена! Выберите нужное действие в меню ниже \u2B07\uFE0F",
		"btn_support":       "Поддержка",
		"btn_about":         "О нас",
		"btn_contacts":      "Контакты",
		"btn_settings":      "Настройки",
		"btn_back":          "🔙 Назад",
		"btn_finish_chat":   "❌ Завершить чат",
		"lang_changed":      "Вы успешно изменили язык.",
		"err_command":       "Просим прощения, данная команда не распознана. Пожалуйста, воспользуйтесь кнопками меню для навигации.",
		"about_text":        "Рады приветствовать Вас! Мы являемся надежным финансовым партнером, предоставляющим современные и безопасные банковские услуги.",
		"contacts_text":     "Уважаемый клиент, Вы всегда можете связаться с нами любым удобным для Вас способом:\n\n📞 Круглосуточный контакт-центр: +998 71 123-45-67\n📧 Электронная почта: info@asakabank.uz\n\nМы всегда рады помочь и ответить на Ваши вопросы.",
		"support_welcome":   "Добро пожаловать в службу заботы о клиентах. Пожалуйста, выберите направление, по которому у Вас возник вопрос, используя меню ниже \u2B07\uFE0F:",
		"ask_service":       "По какой услуге у Вас есть вопросы или возникли проблемы?",
		"wait_operator":     "Вы выбрали услугу: *%s*.\n\nПожалуйста, ожидайте, мы соединяем Вас с первым освободившимся специалистом...",
		"wait_general":      "Благодарим Вас за обращение. Система подготавливает соединение с первым освободившимся оператором по общим вопросам. Пожалуйста, ожидайте...\n\n⚠️ *Примечание:* Вы можете вернуться в меню, нажав кнопку «Назад» ниже.",
		"err_unavailable":   "Приносим извинения, в данный момент сервис временно недоступен. Пожалуйста, попробуйте позже.",
		"chat_finished":     "Диалог завершен. Вы вернулись в главное меню.",
		"no_session":        "К сожалению, активная сессия не найдена. Пожалуйста, начните новый диалог через меню «Поддержка».",
		"back_to_main":      "Вы вернулись в главное меню. Пожалуйста, выберите нужное действие \u2B07\uFE0F",
		"op_panel_title":    "👨‍💻 Рабочее место оператора",
		"op_status_online":  "🟢 Начать смену (Online)",
		"op_status_offline": "🔴 Завершить смену (Offline)",
		"op_stats":          "📊 Моя статистика",
		"op_msg_online":     "✅ Вы вышли на смену. Теперь Вы будете получать новые заявки.",
		"op_msg_offline":    "⏸ Смена завершена. Поступление заявок приостановлено.",

		// База данных (Отделы)
		"Физ. лица":          "Физ. лица",
		"Юр. лица":           "Юр. лица",
		"Махалла банкирлари": "Махалла банкирлари",
		"Общие вопросы":      "Общие вопросы",

		// База данных (Услуги)
		"Кредиты": "Кредиты", "Вклады": "Вклады", "Карты": "Карты",
		"Денежные переводы": "Денежные переводы", "Курс валют": "Курс валют", "Акции": "Акции",
		"Запись онлайн": "Запись онлайн", "Тарифы": "Тарифы", "Asaka Travel": "Asaka Travel",
		"ESG": "ESG", "Депозиты": "Депозиты", "Финансирование": "Финансирование",
		"Эквайринг": "Эквайринг", "Фармацевтика": "Фармацевтика",
		"Кредитные линии": "Кредитные линии", "Интернет банкинг": "Интернет банкинг",
	},
	"uz": {
		"choose_lang":       "Iltimos, xizmat ko'rsatish tilini tanlang:",
		"ask_name":          "Assalomu alaykum! Iltimos, Ism va Familiyangizni yozing.\nMisol: Ivan Ivanov",
		"btn_share_contact": "📱 Kontaktni ulashish",
		"ask_phone":         "Ajoyib! Endi telefon raqamingizni yuboring yoki yozing.\nMisol: +998901234567",
		"reg_success":       "Ro'yxatdan o'tish muvaffaqiyatli yakunlandi! Quyidagi menyudan kerakli amalni tanlang \u2B07\uFE0F",
		"btn_support":       "Qo'llab-quvvatlash",
		"btn_about":         "Biz haqimizda",
		"btn_contacts":      "Aloqa",
		"btn_settings":      "Sozlamalar",
		"btn_back":          "🔙 Orqaga",
		"btn_finish_chat":   "❌ Chatni yakunlash",
		"lang_changed":      "Til muvaffaqiyatli o'zgartirildi.",
		"err_command":       "Kechirasiz, buyruq taninmadi. Iltimos, menyu tugmalaridan foydalaning.",
		"about_text":        "Sizni kutib olishdan xursandmiz! Biz zamonaviy va xavfsiz bank xizmatlarini taqdim etuvchi ishonchli moliyaviy hamkormiz.",
		"contacts_text":     "Hurmatli mijoz, Siz biz bilan o'zingizga qulay bo'lgan har qanday usulda bog'lanishingiz mumkin:\n\n📞 Kecha-yu kunduz ishlaydigan aloqa markazi: +998 71 123-45-67\n📧 Elektron pochta: info@asakabank.uz\n\nBiz doimo yordam berishga va savollaringizga javob berishga tayyormiz.",
		"support_welcome":   "Mijozlarga g'amxo'rlik qilish xizmatiga xush kelibsiz. Iltimos, quyidagi menyudan foydalanib, savolingiz bor yo'nalishni tanlang \u2B07\uFE0F:",
		"ask_service":       "Qaysi xizmat bo'yicha savollaringiz bor yoki muammolarga duch keldingiz?",
		"wait_operator":     "Siz quyidagi xizmatni tanladingiz: *%s*.\n\nIltimos kuting, biz Sizni birinchi bo'sh mutaxassis bilan bog'laymiz...",
		"wait_general":      "Murojaatingiz uchun tashakkur. Tizim birinchi bo'sh bo'lgan umumiy savollar bo'yicha operator bilan ulanishni tayyorlamoqda. Iltimos kuting...\n\n⚠️ *Eslatma:* Pastdagi «Orqaga» tugmasini bosib menyuga qaytishingiz mumkin.",
		"err_unavailable":   "Kechirasiz, ayni damda xizmat vaqtincha mavjud emas. Iltimos, keyinroq qayta urinib ko'ring.",
		"chat_finished":     "Muloqot yakunlandi. Siz asosiy menyuga qaytdingiz.",
		"no_session":        "Afsuski, faol sessiya topilmadi. Iltimos, «Qo'llab-quvvatlash» menyusi orqali yangi muloqotni boshlang.",
		"back_to_main":      "Siz asosiy menyuga qaytdingiz. Iltimos, quyidagi klaviaturadan kerakli amalni tanlang \u2B07\uFE0F",
		"op_panel_title":    "👨‍💻 Operator ish joyi",
		"op_status_online":  "🟢 Ishni boshlash(online)",
		"op_status_offline": "🔴 Ishni tugatish(offline)",
		"op_stats":          "📊 Mening statistikam",
		"op_msg_online":     "✅ Siz ishga tushdingiz. Endi yangi arizalarni qabul qilasiz.",
		"op_msg_offline":    "⏸ Ish tugatildi. Arizalar qabul qilinmayapti.",

		// База данных (Отделы)
		"Физ. лица":          "Jismoniy shaxslar",
		"Юр. лица":           "Yuridik shaxslar",
		"Махалла банкирлари": "Mahalla bankirlari",
		"Общие вопросы":      "Umumiy savollar",

		// База данных (Услуги)
		"Кредиты": "Kreditlar", "Вклады": "Omonatlar", "Карты": "Kartalar",
		"Денежные переводы": "Pul o'tkazmalari", "Курс валют": "Valyuta kurslari", "Акции": "Aksiyalar",
		"Запись онлайн": "Onlayn yozilish", "Тарифы": "Tariflar", "Asaka Travel": "Asaka Travel",
		"ESG": "ESG", "Депозиты": "Depozitlar", "Финансирование": "Moliyalashtirish",
		"Эквайринг": "Ekvayring", "Фармацевтика": "Farmatsevtika",
		"Кредитные линии": "Kredit liniyalari", "Интернет банкинг": "Internet banking",
	},
}

// Get возвращает переведенный текст. Если ключа нет, возвращает сам ключ.
func Get(lang string, key string) string {
	if lang == "" {
		lang = "ru"
	}
	if text, ok := messages[lang][key]; ok {
		return text
	}
	return key
}
