package services

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/gomail.v2"
)

/*
Файл email.go содержит функции для отправки пользователям
транзакционных писем через SMTP: подтверждение регистрации,
уведомление об удалении неверифицированного аккаунта и код
для сброса пароля. Каждое письмо оформлено в виде HTML-шаблона
*/

/*
SendEmail отправляет пользователю письмо со ссылкой
для подтверждения адреса электронной почты и обновляет
время последней отправки письма
*/
func SendEmail(db *pgxpool.Pool, email string) error {
	var verificationCode string

	err := db.QueryRow(
		context.Background(),
		`SELECT verification_code FROM users WHERE email = $1`,
		email,
	).Scan(&verificationCode)

	if err != nil {
		return err
	}

	_, err = db.Exec(
		context.Background(),
		`UPDATE users SET last_send = $1
		WHERE email = $2`,
		time.Now().UTC(), email)

	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_USER"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Подтверждение почты для LinkShorter")
	m.SetBody("text/html", `
	<!DOCTYPE html>
	<html lang="ru">
	<head>
		<meta charset="UTF-8">
	</head>
	<body style="margin:0;padding:0;background:#f4f4f7;font-family:Arial,Helvetica,sans-serif;">
		<table width="100%" cellpadding="0" cellspacing="0" style="background:#f4f4f7;padding:40px 0;">
			<tr>
				<td align="center">
					<table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:12px;padding:40px;box-shadow:0 4px 16px rgba(0,0,0,.08);">
						<tr>
							<td align="center">
								<h1 style="margin:0;color:#333;font-size:28px;">
									Подтверждение почты
								</h1>

								<p style="margin:24px 0 16px;color:#555;font-size:16px;line-height:1.6;">
									Спасибо за регистрацию! Для подтверждения адреса электронной почты
									нажмите на кнопку ниже
								</p>

                            	<a href="http://localhost:8080/api/verify/`+verificationCode+`"
                               		style="display:inline-block;background:#2563eb;color:#ffffff;
                                    	  text-decoration:none;padding:14px 32px;
                                    	  border-radius:8px;font-size:16px;font-weight:bold;">
                                	Подтвердить почту
                            	</a>

								<p style="margin:24px 0 0;color:#777;font-size:14px;">
									Ссылка действительна <strong>3 дня</strong>
								</p>

                            	<hr style="margin:32px 0;border:none;border-top:1px solid #e5e7eb;">

								<p style="margin:0;color:#999;font-size:13px;line-height:1.5;">
									Если вы не регистрировались на нашем сайте, просто проигнорируйте это письмо
								</p>
                        	</td>
                    	</tr>
                	</table>
            	</td>
       	 </tr>
    	</table>
	</body>
	</html>
	`)

	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	d := gomail.NewDialer(
		os.Getenv("SMTP_HOST"),
		port,
		os.Getenv("SMTP_USER"),
		os.Getenv("SMTP_PASSWORD"),
	)

	return d.DialAndSend(m)
}

/*
SendAccountDeletingEmail отправляет пользователю письмо
с уведомлением об автоматическом удалении аккаунта
из-за неподтвержденной электронной почты
*/
func SendAccountDeletingEmail(email string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_USER"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Аккаунт LinkShorter был удалён")

	m.SetBody("text/html", `
	<!DOCTYPE html>
	<html lang="ru">
	<head>
		<meta charset="UTF-8">
	</head>
	<body style="margin:0;padding:0;background:#f4f4f7;font-family:Arial,Helvetica,sans-serif;">
		<table width="100%" cellpadding="0" cellspacing="0" style="background:#f4f4f7;padding:40px 0;">
			<tr>
				<td align="center">
					<table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:12px;padding:40px;box-shadow:0 4px 16px rgba(0,0,0,.08);">
						<tr>
							<td align="center">
	
								<h1 style="margin:0;color:#333;font-size:28px;">
									Аккаунт удалён
								</h1>
	
								<p style="margin:24px 0 16px;color:#555;font-size:16px;line-height:1.6;">
									К сожалению, Вы не подтвердили адрес электронной почты
									в течение <strong>3 дней</strong>, поэтому Ваш аккаунт был автоматически удалён...
								</p>

								<div style="display:inline-block;background:#f3f4f6;padding:16px 24px;border-radius:8px;color:#444;font-size:15px;line-height:1.6;">
									Нам очень жаль, что так получилось
								</div>
	
								<p style="margin:24px 0 0;color:#555;font-size:16px;line-height:1.6;">
									Если Вы всё ещё хотите пользоваться <strong>LinkShorter</strong>,
									просто зарегистрируйтесь снова и подтвердите адрес электронной почты
								</p>
	
								<hr style="margin:32px 0;border:none;border-top:1px solid #e5e7eb;">
	
								<p style="margin:0;color:#999;font-size:13px;line-height:1.5;">
									Спасибо, что проявили интерес к LinkShorter. Будем рады видеть Вас снова!
								</p>
	
							</td>
						</tr>
					</table>
				</td>
			</tr>
		</table>
	</body>
	</html>
	`)

	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	d := gomail.NewDialer(
		os.Getenv("SMTP_HOST"),
		port,
		os.Getenv("SMTP_USER"),
		os.Getenv("SMTP_PASSWORD"),
	)

	return d.DialAndSend(m)
}

/*
SendPasswordResetEmail отправляет пользователю письмо
с шестизначным кодом подтверждения для смены
пароля аккаунта
*/
func SendPasswordResetEmail(email string, code string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_USER"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Код подтверждения LinkShorter")
	m.SetBody("text/html", `
	<!DOCTYPE html>
	<html lang="ru">
	<head>
		<meta charset="UTF-8">
	</head>
	<body style="margin:0;padding:0;background:#f4f4f7;font-family:Arial,Helvetica,sans-serif;">
		<table width="100%" cellpadding="0" cellspacing="0" style="background:#f4f4f7;padding:40px 0;">
			<tr>
				<td align="center">
					<table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:12px;padding:40px;box-shadow:0 4px 16px rgba(0,0,0,.08);">
						<tr>
							<td align="center">
								<h1 style="margin:0;color:#333;font-size:28px;">
									Подтверждение почты
								</h1>
								<p style="margin:24px 0 16px;color:#555;font-size:16px;line-height:1.6;">
									Для подтверждения адреса электронной почты,
									чтобы сменить пароль от аккаунта, используйте следующий код
								</p>
								<div style="
									display:inline-block;
									background:#2563eb;
									color:#ffffff;
									font-size:34px;
									font-weight:bold;
									letter-spacing:8px;
									padding:18px 36px;
									border-radius:10px;
									margin:16px 0;">
									`+code+`
								</div>
								<p style="margin:24px 0 0;color:#777;font-size:14px;">
									Код действителен <strong>15 минут</strong>
								</p>
								<hr style="margin:32px 0;border:none;border-top:1px solid #e5e7eb;">
								<p style="margin:0;color:#999;font-size:13px;line-height:1.5;">
									Если Вы не запрашивали этот код, просто проигнорируйте это письмо
								</p>
							</td>
						</tr>
					</table>
				</td>
			</tr>
		</table>
	</body>
	</html>
	`)

	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	d := gomail.NewDialer(
		os.Getenv("SMTP_HOST"),
		port,
		os.Getenv("SMTP_USER"),
		os.Getenv("SMTP_PASSWORD"),
	)
	return d.DialAndSend(m)
}
