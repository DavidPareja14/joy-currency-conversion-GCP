import functions_framework
from flask import jsonify
import smtplib
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
import os

@functions_framework.http
def send_email_notification(request):
    request_json = request.get_json(silent=True)
    
    if not request_json:
        return jsonify({"error": "invalid request body"}), 400
    
    email = request_json.get('email')
    currency_origin = request_json.get('currency_origin')
    currency_destination = request_json.get('currency_destination')
    current_rate = request_json.get('current_rate')
    threshold = request_json.get('threshold')
    
    if not all([email, currency_origin, currency_destination, current_rate, threshold]):
        return jsonify({"error": "missing required fields"}), 400
    
    subject = f"Currency Alert: {currency_origin} to {currency_destination}"
    body = f"""
    Hello,
    
    Your currency conversion threshold has been reached!
    
    Details:
    - From: {currency_origin}
    - To: {currency_destination}
    - Current Rate: {current_rate}
    - Your Threshold: {threshold}
    
    The current rate is now above your configured threshold.
    
    Best regards,
    Currency Conversion Service
    """
    
    try:
        send_email(email, subject, body)
        return jsonify({"message": "email sent successfully", "status": "success"}), 200
    except Exception as e:
        print(f"Error sending email: {e}")
        return jsonify({"error": str(e), "status": "error"}), 500

def send_email(to_email, subject, body):
    smtp_server = os.environ.get('SMTP_SERVER', 'smtp.gmail.com')
    smtp_port = int(os.environ.get('SMTP_PORT', '587'))
    smtp_user = os.environ.get('SMTP_USER')
    smtp_password = os.environ.get('SMTP_PASSWORD')
    
    if not smtp_user or not smtp_password:
        raise ValueError("SMTP credentials not configured")
    
    msg = MIMEMultipart()
    msg['From'] = smtp_user
    msg['To'] = to_email
    msg['Subject'] = subject
    
    msg.attach(MIMEText(body, 'plain'))
    
    server = smtplib.SMTP(smtp_server, smtp_port)
    server.starttls()
    server.login(smtp_user, smtp_password)
    text = msg.as_string()
    server.sendmail(smtp_user, to_email, text)
    server.quit()