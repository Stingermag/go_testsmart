-- phpMyAdmin SQL Dump
-- version 4.9.0.1
-- https://www.phpmyadmin.net/
--
-- Хост: localhost
-- Время создания: Сен 16 2020 г., 18:55
-- Версия сервера: 5.7.14
-- Версия PHP: 7.0.9

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET AUTOCOMMIT = 0;
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- База данных: `go_testsmart_user`
--

-- --------------------------------------------------------

--
-- Структура таблицы `users`
--

CREATE TABLE `users` (
  `ID` int(10) NOT NULL,
  `name` varchar(100) NOT NULL,
  `surname` varchar(100) NOT NULL,
  `age` varchar(100) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Дамп данных таблицы `users`
--

INSERT INTO `users` (`ID`, `name`, `surname`, `age`) VALUES
(1, 'dima', 'treert', '4'),
(2, 'dimas', 'koles', '56'),
(3, 'Dmitry', 'Ezhov', '66'),
(4, 'Nikolai', 'Tesla', '36'),
(5, 'Sergey', 'Penkin', '76'),
(6, 'Dmitry', 'DonScoi', '34'),
(7, 'Dmitry', 'DonScoi', '54'),
(8, 'Dmitry', 'DonScoi', '43'),
(9, 'Dmitry', 'DonScoi', '34'),
(10, 'Dmitry', 'DonScoi', '34'),
(11, 'Dmitry', 'DonScoi', '34'),
(12, 'Dmitry', 'DonScoi', '31'),
(13, 'Dmitry', 'DonScoi', '31'),
(14, 'Dmitry', 'Donscoi', '34'),

--
-- Индексы сохранённых таблиц
--

--
-- Индексы таблицы `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`ID`);

--
-- AUTO_INCREMENT для сохранённых таблиц
--

--
-- AUTO_INCREMENT для таблицы `users`
--
ALTER TABLE `users`
  MODIFY `ID` int(10) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=149;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
