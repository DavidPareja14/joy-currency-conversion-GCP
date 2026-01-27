import functions_framework
from flask import jsonify
import smtplib
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
import os
import base64
import json

@functions_framework.http
def send_email_notification(request):
    request_json = request.get_json(silent=True)
    
    if not request_json:
        return jsonify({"error": "invalid request body"}), 400
    
    # Extract message from Pub/Sub format
    if 'message' in request_json:
        # Pub/Sub Push format
        pubsub_message = request_json['message']
        if 'data' not in pubsub_message:
            return jsonify({"error": "no data in pubsub message"}), 400
        
        # Decode base64 message
        data_base64 = pubsub_message['data']
        data_decoded = base64.b64decode(data_base64).decode('utf-8')
        data = json.loads(data_decoded)
    else:
        # Direct HTTP call (backward compatibility)
        data = request_json
    
    email = data.get('email')
    currency_origin = data.get('currency_origin')
    currency_destination = data.get('currency_destination')
    current_rate = data.get('current_rate')
    threshold = data.get('threshold')
    
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