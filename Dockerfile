FROM python:3.6-slim

WORKDIR /bashpush

COPY requirements.txt .
RUN pip install -r requirements.txt

COPY run.sh .
COPY bashpush.py .

CMD ["./run.sh"]
